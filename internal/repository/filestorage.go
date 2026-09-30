package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	models "github.com/TimurBattalkhanov/metrics-collector/internal/model"
	"go.uber.org/zap"
)

type FileStorage struct {
	*MemStorage
	path     string
	syncSave bool
}

func NewFileStorage(storage *MemStorage, path string, syncSave bool) *FileStorage {
	return &FileStorage{
		path:       path,
		syncSave:   syncSave,
		MemStorage: storage,
	}
}

func (fs *FileStorage) RunIntervalSave(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := fs.Save(); err != nil {
				zap.S().Errorw("failed to save metrics", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (fs *FileStorage) Save() error {
	gauges := fs.MemStorage.GetGauges()
	counters := fs.MemStorage.GetCounters()
	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))
	for k, v := range gauges {
		metric := models.Metrics{
			ID:    k,
			MType: models.Gauge,
			Value: &v,
		}
		metrics = append(metrics, metric)
	}
	for k, v := range counters {
		metric := models.Metrics{
			ID:    k,
			MType: models.Counter,
			Delta: &v,
		}
		metrics = append(metrics, metric)
	}

	producer, err := NewProducer(fs.path)
	if err != nil {
		return err
	}
	defer producer.Close()

	err = producer.WriteMetrics(metrics)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) Load() error {
	consumer, err := NewConsumer(fs.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	defer consumer.Close()

	metrics, err := consumer.ReadMetrics()
	if err != nil {
		return err
	}

	for _, v := range metrics {
		if v.MType == models.Gauge && v.Value != nil {
			fs.MemStorage.UpdateGauge(v.ID, *v.Value)
			continue
		}
		if v.MType == models.Counter && v.Delta != nil {
			fs.MemStorage.UpdateCounter(v.ID, *v.Delta)
		}
	}
	return nil
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteMetrics(metrics []models.Metrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadMetrics() ([]models.Metrics, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	var metrics []models.Metrics
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (fs *FileStorage) UpdateGauge(id string, value float64) float64 {
	v := fs.MemStorage.UpdateGauge(id, value)
	if fs.syncSave {
		if err := fs.Save(); err != nil {
			zap.S().Errorf("failed to save metrics: %s", err.Error())
		}
	}
	return v
}

func (fs *FileStorage) UpdateCounter(id string, value int64) int64 {
	v := fs.MemStorage.UpdateCounter(id, value)
	if fs.syncSave {
		if err := fs.Save(); err != nil {
			zap.S().Errorf("failed to save metrics: %s", err.Error())
		}
	}
	return v
}
