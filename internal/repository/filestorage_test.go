package repository

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileStorageSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")

	store := NewFileStorage(NewMemStorage(), path, false)
	store.UpdateGauge(ctx, "Alloc", 12.5)
	store.UpdateCounter(ctx, "PollCount", 3)
	store.UpdateCounter(ctx, "PollCount", 2)

	if err := store.Save(ctx); err != nil {
		t.Fatal(err)
	}

	restored := NewFileStorage(NewMemStorage(), path, false)
	if err := restored.Load(ctx); err != nil {
		t.Fatal(err)
	}

	wantGauges := map[string]float64{"Alloc": 12.5}
	wantCounters := map[string]int64{"PollCount": 5}

	if got, _ := restored.GetGauges(ctx); !reflect.DeepEqual(got, wantGauges) {
		t.Errorf("gauges: получили %v, ожидали %v", got, wantGauges)
	}
	if got, _ := restored.GetCounters(ctx); !reflect.DeepEqual(got, wantCounters) {
		t.Errorf("counters: получили %v, ожидали %v", got, wantCounters)
	}
}

func TestFileStorageLoad(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name         string
		createFile   bool
		content      string
		wantErr      bool
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			"Valid metrics",
			true,
			`[{"id":"Alloc","type":"gauge","value":1.5},{"id":"PollCount","type":"counter","delta":7}]`,
			false,
			map[string]float64{"Alloc": 1.5},
			map[string]int64{"PollCount": 7},
		},
		{
			"No file",
			false,
			"",
			false,
			map[string]float64{},
			map[string]int64{},
		},
		{
			"Empty file",
			true,
			"",
			false,
			map[string]float64{},
			map[string]int64{},
		},
		{
			"Entries without value are skipped",
			true,
			`[{"id":"X","type":"gauge"},{"id":"Y","type":"counter"},{"id":"W","type":"gauge","delta":1},{"id":"Z","type":"gauge","value":2}]`,
			false,
			map[string]float64{"Z": 2},
			map[string]int64{},
		},
		{
			"Invalid JSON",
			true,
			`not a json`,
			true,
			map[string]float64{},
			map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metrics.json")
			if tt.createFile {
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			store := NewFileStorage(NewMemStorage(), path, false)
			err := store.Load(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка: получили %v, ожидали ошибку: %v", err, tt.wantErr)
			}
			if got, _ := store.GetGauges(ctx); !reflect.DeepEqual(got, tt.wantGauges) {
				t.Errorf("gauges: получили %v, ожидали %v", got, tt.wantGauges)
			}
			if got, _ := store.GetCounters(ctx); !reflect.DeepEqual(got, tt.wantCounters) {
				t.Errorf("counters: получили %v, ожидали %v", got, tt.wantCounters)
			}
		})
	}
}

func TestFileStorageLoadDoesNotRewriteFile(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")
	content := `[{"id":"Alloc","type":"gauge","value":1.5}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	store := NewFileStorage(NewMemStorage(), path, true)
	if err := store.Load(ctx); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("файл: получили %q, ожидали %q", data, content)
	}
}

func TestFileStorageSyncSave(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		syncSave bool
		update   func(s *FileStorage)
		wantFile string
	}{
		{
			"Sync gauge",
			true,
			func(s *FileStorage) { s.UpdateGauge(ctx, "Alloc", 12.5) },
			`[{"id":"Alloc","type":"gauge","value":12.5}]` + "\n",
		},
		{
			"Sync counter",
			true,
			func(s *FileStorage) { s.UpdateCounter(ctx, "PollCount", 3) },
			`[{"id":"PollCount","type":"counter","delta":3}]` + "\n",
		},
		{
			"Without sync file is not created",
			false,
			func(s *FileStorage) { s.UpdateGauge(ctx, "Alloc", 12.5) },
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metrics.json")

			store := NewFileStorage(NewMemStorage(), path, tt.syncSave)
			tt.update(store)

			data, err := os.ReadFile(path)
			if tt.wantFile == "" {
				if !os.IsNotExist(err) {
					t.Errorf("файл не должен был создаться, ошибка: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tt.wantFile {
				t.Errorf("файл: получили %q, ожидали %q", data, tt.wantFile)
			}
		})
	}
}
