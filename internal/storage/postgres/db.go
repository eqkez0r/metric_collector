// Пакет postgres представляет реализацию хранилища на основе
// PostgreSQL
package postgres

import (
	"context"

	store "github.com/Eqke/metric-collector/internal/storage"
	"github.com/Eqke/metric-collector/pkg/metric"
	"github.com/Eqke/metric-collector/utils/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	// Тип хранилища
	TYPE = "PostgresSQL database"

	// Запросы на создание таблиц
	queryCreateGauges   = `CREATE TABLE IF NOT EXISTS gauges(name text primary key, value double precision)`
	queryCreateCounters = `CREATE TABLE IF NOT EXISTS counters(name text primary key, value bigint)`

	// Запросы для gauge
	queryGetGauge    = `SELECT value FROM gauges WHERE name = $1 LIMIT 1`
	queryGetAllGauge = `SELECT name, value FROM gauges`
	querySetGauge    = `INSERT INTO gauges(name, value) VALUES($1, $2) ON CONFLICT(name) DO UPDATE SET value = $2`

	// Запросы для counter
	queryGetCounter    = `SELECT value FROM counters WHERE name = $1 LIMIT 1`
	queryGetAllCounter = `SELECT name, value FROM counters`
	querySetCounter    = `INSERT INTO counters(name, value) VALUES($1, $2) ON CONFLICT(name) DO UPDATE SET value = counters.value + EXCLUDED.value`
)

// Тип PSQLStorage представляет реализацию хранилища
type PSQLStorage struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

// Функция New возвращает экземпляр хранилища PSQLStorage
func New(ctx context.Context, logger *zap.SugaredLogger, conn string) (*PSQLStorage, error) {

	db, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}

	err = retry.Retry(logger, 3, func() error {
		err = db.Ping(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = retry.Retry(logger, 3, func() error {
		_, err = db.Exec(ctx, queryCreateGauges)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = retry.Retry(logger, 3, func() error {
		_, err = db.Exec(ctx, queryCreateCounters)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PSQLStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (p *PSQLStorage) SetValue(ctx context.Context, metricType, name, value string) error {
	switch metricType {
	case metric.TypeCounter.String():
		{
			_, err := p.db.Exec(ctx, querySetCounter, name, value)
			if err != nil {
				p.logger.Errorf("Database exec error: %v", err)
				return err
			}
		}
	case metric.TypeGauge.String():
		{
			_, err := p.db.Exec(ctx, querySetGauge, name, value)
			if err != nil {
				p.logger.Errorf("Database exec error: %v", err)
				return err
			}
		}
	default:
		{
			p.logger.Error(store.ErrPointSetValue, store.ErrIsUnknownType)
			return store.ErrIsUnknownType

		}
	}
	p.logger.Infof("metric was saved with type: %s, name: %s, value: %s",
		metricType, name, value)

	return nil
}

func (p *PSQLStorage) SetMetric(ctx context.Context, m metric.Metrics) error {
	switch m.MType {
	case metric.TypeCounter.String():
		{
			_, err := p.db.Exec(ctx, querySetCounter, m.ID, *m.Delta)
			if err != nil {
				p.logger.Errorf("Database exec error: %v. metric: %v", err, m)
				return err
			}
		}
	case metric.TypeGauge.String():
		{
			_, err := p.db.Exec(ctx, querySetGauge, m.ID, *m.Value)
			if err != nil {
				p.logger.Errorf("Database exec error: %v. metric: %v", err, m)
				return err
			}
		}
	default:
		{
			p.logger.Error(store.ErrPointSetValue, store.ErrIsUnknownType)
			return store.ErrIsUnknownType
		}
	}
	return nil
}

func (p *PSQLStorage) SetMetrics(ctx context.Context, m []metric.Metrics) error {
	if len(m) == 0 {
		return nil
	}
	for _, v := range m {
		err := p.SetMetric(ctx, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PSQLStorage) GetValue(ctx context.Context, metricType, name string) (string, error) {
	var row pgx.Row
	var value string
	switch metricType {
	case metric.TypeCounter.String():
		{
			row = p.db.QueryRow(ctx, queryGetCounter, name)
			if err := row.Scan(&value); err != nil {
				p.logger.Errorf("Database scan metricType: %s, name: %s error: %v. ", metricType, name, err)
				return "", err
			}
			return value, nil
		}
	case metric.TypeGauge.String():
		{
			row = p.db.QueryRow(ctx, queryGetGauge, name)
			if err := row.Scan(&value); err != nil {
				p.logger.Errorf("Database scan metricType: %s, name: %s error: %v.", metricType, name, err)
				return "", err
			}
			return value, nil
		}
	default:
		{
			p.logger.Error(store.ErrPointGetValue, store.ErrIsUnknownType)
			return "", store.ErrIsUnknownType
		}
	}
}

func (p *PSQLStorage) GetMetrics(ctx context.Context) (map[string][]store.Metric, error) {
	metrics := make(map[string][]store.Metric, 2)
	metrics[metric.TypeCounter.String()] = make([]store.Metric, 0, 2)
	metrics[metric.TypeGauge.String()] = make([]store.Metric, 0, 31)
	var rows pgx.Rows
	var err error

	rows, err = p.db.Query(ctx, queryGetAllGauge)
	if err != nil {
		p.logger.Errorf("Database query error: %v", err)
		return nil, err
	}

	for rows.Next() {
		var m store.Metric
		if err = rows.Scan(&m.Name, &m.Value); err != nil {
			p.logger.Errorf("Database scan metric: %v error: %v. ", m, err)
			return nil, err
		}
		metrics[metric.TypeGauge.String()] = append(metrics[metric.TypeGauge.String()], m)
	}

	rows, err = p.db.Query(ctx, queryGetAllCounter)
	if err != nil {
		p.logger.Errorf("Database query error: %v", err)
		return nil, err
	}

	for rows.Next() {
		var m store.Metric
		if err = rows.Scan(&m.Name, &m.Value); err != nil {
			p.logger.Errorf("Database scan metric: %v error: %v. ", m, err)
			return nil, err
		}
		metrics[metric.TypeCounter.String()] = append(metrics[metric.TypeCounter.String()], m)
	}

	return metrics, nil
}

func (p *PSQLStorage) GetMetric(ctx context.Context, m metric.Metrics) (metric.Metrics, error) {

	switch m.MType {
	case metric.TypeCounter.String():
		{
			if err := p.db.QueryRow(ctx, queryGetCounter, m.ID).Scan(&m.Delta); err != nil {
				p.logger.Errorf("Database scan metric: %v error: %v. ", m, err)
				return m, err
			}

		}
	case metric.TypeGauge.String():
		{
			if err := p.db.QueryRow(ctx, queryGetGauge, m.ID).Scan(&m.Value); err != nil {
				p.logger.Errorf("Database scan metric: %v error: %v. ", m, err)
				return m, err
			}

		}
	default:
		{
			p.logger.Error(store.ErrPointGetMetric, store.ErrIsUnknownType)
			return m, store.ErrIsUnknownType
		}
	}
	p.logger.Infof("metric %v", m)
	return m, nil
}

func (p *PSQLStorage) ToJSON(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (p *PSQLStorage) FromJSON(ctx context.Context, bytes []byte) error {
	return nil
}

func (p *PSQLStorage) ToFile(ctx context.Context, s string) error {
	return nil
}

func (p *PSQLStorage) FromFile(ctx context.Context, s string) error {
	return nil
}

func (p *PSQLStorage) Close() error {
	p.db.Close()
	return nil
}

func (p *PSQLStorage) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}

func (p *PSQLStorage) Type() string {
	return TYPE
}
