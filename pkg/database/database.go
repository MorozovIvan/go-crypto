package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// MarketData represents a market data entry
type MarketData struct {
	ID        int64     `json:"id"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Indicator string    `json:"indicator"`
	Score     float64   `json:"score"`
	ChartData string    `json:"chart_data"` // JSON string
	Timestamp time.Time `json:"timestamp"`
}

// SignalHistory represents historical signal data
type SignalHistory struct {
	ID         int64     `json:"id"`
	Signal     string    `json:"signal"`
	Confidence string    `json:"confidence"`
	Score      float64   `json:"score"`
	Metrics    string    `json:"metrics"` // JSON string of all metrics
	Timestamp  time.Time `json:"timestamp"`
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	db := &DB{conn: conn}

	// Initialize tables
	if err := db.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return db, nil
}

// initTables creates the necessary database tables
func (db *DB) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS market_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric TEXT NOT NULL,
			value REAL NOT NULL,
			indicator TEXT NOT NULL,
			score REAL NOT NULL,
			chart_data TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_market_data_metric ON market_data(metric)`,
		`CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data(timestamp)`,
		`CREATE TABLE IF NOT EXISTS signal_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			signal TEXT NOT NULL,
			confidence TEXT NOT NULL,
			score REAL NOT NULL,
			metrics TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_history_timestamp ON signal_history(timestamp)`,
		`CREATE TABLE IF NOT EXISTS api_performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			api_name TEXT NOT NULL,
			response_time INTEGER NOT NULL,
			success BOOLEAN NOT NULL,
			error_message TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_performance_api_name ON api_performance(api_name)`,
		`CREATE INDEX IF NOT EXISTS idx_api_performance_timestamp ON api_performance(timestamp)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return nil
}

// StoreMarketData stores market data in the database
func (db *DB) StoreMarketData(data *MarketData) error {
	chartDataJSON, err := json.Marshal(data.ChartData)
	if err != nil {
		return fmt.Errorf("failed to marshal chart data: %v", err)
	}

	query := `INSERT INTO market_data (metric, value, indicator, score, chart_data, timestamp) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err = db.conn.Exec(query, data.Metric, data.Value, data.Indicator,
		data.Score, string(chartDataJSON), data.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to store market data: %v", err)
	}

	return nil
}

// GetMarketDataHistory retrieves historical market data for a metric
func (db *DB) GetMarketDataHistory(metric string, hours int) ([]*MarketData, error) {
	query := `SELECT id, metric, value, indicator, score, chart_data, timestamp 
			  FROM market_data 
			  WHERE metric = ? AND timestamp > datetime('now', '-' || ? || ' hours')
			  ORDER BY timestamp DESC`

	rows, err := db.conn.Query(query, metric, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data: %v", err)
	}
	defer rows.Close()

	var results []*MarketData
	for rows.Next() {
		var data MarketData
		var chartDataJSON string

		err := rows.Scan(&data.ID, &data.Metric, &data.Value, &data.Indicator,
			&data.Score, &chartDataJSON, &data.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Parse chart data JSON
		if chartDataJSON != "" {
			if err := json.Unmarshal([]byte(chartDataJSON), &data.ChartData); err != nil {
				log.Printf("Failed to unmarshal chart data for %s: %v", metric, err)
			}
		}

		results = append(results, &data)
	}

	return results, nil
}

// StoreSignalHistory stores signal history in the database
func (db *DB) StoreSignalHistory(signal *SignalHistory) error {
	query := `INSERT INTO signal_history (signal, confidence, score, metrics, timestamp) 
			  VALUES (?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, signal.Signal, signal.Confidence,
		signal.Score, signal.Metrics, signal.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to store signal history: %v", err)
	}

	return nil
}

// GetSignalHistory retrieves historical signals
func (db *DB) GetSignalHistory(hours int) ([]*SignalHistory, error) {
	query := `SELECT id, signal, confidence, score, metrics, timestamp 
			  FROM signal_history 
			  WHERE timestamp > datetime('now', '-' || ? || ' hours')
			  ORDER BY timestamp DESC`

	rows, err := db.conn.Query(query, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query signal history: %v", err)
	}
	defer rows.Close()

	var results []*SignalHistory
	for rows.Next() {
		var signal SignalHistory

		err := rows.Scan(&signal.ID, &signal.Signal, &signal.Confidence,
			&signal.Score, &signal.Metrics, &signal.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		results = append(results, &signal)
	}

	return results, nil
}

// StoreAPIPerformance stores API performance metrics
func (db *DB) StoreAPIPerformance(apiName string, responseTime int, success bool, errorMessage string) error {
	query := `INSERT INTO api_performance (api_name, response_time, success, error_message, timestamp) 
			  VALUES (?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, apiName, responseTime, success, errorMessage, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store API performance: %v", err)
	}

	return nil
}

// GetAPIPerformanceStats retrieves API performance statistics
func (db *DB) GetAPIPerformanceStats(apiName string, hours int) (map[string]interface{}, error) {
	query := `SELECT 
				COUNT(*) as total_requests,
				AVG(response_time) as avg_response_time,
				MIN(response_time) as min_response_time,
				MAX(response_time) as max_response_time,
				SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_requests,
				SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_requests
			  FROM api_performance 
			  WHERE api_name = ? AND timestamp > datetime('now', '-' || ? || ' hours')`

	row := db.conn.QueryRow(query, apiName, hours)

	var stats struct {
		TotalRequests      int     `json:"total_requests"`
		AvgResponseTime    float64 `json:"avg_response_time"`
		MinResponseTime    int     `json:"min_response_time"`
		MaxResponseTime    int     `json:"max_response_time"`
		SuccessfulRequests int     `json:"successful_requests"`
		FailedRequests     int     `json:"failed_requests"`
	}

	err := row.Scan(&stats.TotalRequests, &stats.AvgResponseTime, &stats.MinResponseTime,
		&stats.MaxResponseTime, &stats.SuccessfulRequests, &stats.FailedRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to scan performance stats: %v", err)
	}

	successRate := 0.0
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
	}

	return map[string]interface{}{
		"api_name":            apiName,
		"total_requests":      stats.TotalRequests,
		"avg_response_time":   stats.AvgResponseTime,
		"min_response_time":   stats.MinResponseTime,
		"max_response_time":   stats.MaxResponseTime,
		"successful_requests": stats.SuccessfulRequests,
		"failed_requests":     stats.FailedRequests,
		"success_rate":        successRate,
	}, nil
}

// CleanupOldData removes old data to prevent database bloat
func (db *DB) CleanupOldData(daysToKeep int) error {
	queries := []string{
		`DELETE FROM market_data WHERE timestamp < datetime('now', '-' || ? || ' days')`,
		`DELETE FROM signal_history WHERE timestamp < datetime('now', '-' || ? || ' days')`,
		`DELETE FROM api_performance WHERE timestamp < datetime('now', '-' || ? || ' days')`,
	}

	for _, query := range queries {
		result, err := db.conn.Exec(query, daysToKeep)
		if err != nil {
			return fmt.Errorf("failed to cleanup old data: %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		log.Printf("Cleaned up %d old records", rowsAffected)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Global database instance
var globalDB *DB

// InitDB initializes the global database
func InitDB(dbPath string) error {
	var err error
	globalDB, err = NewDB(dbPath)
	if err != nil {
		return err
	}

	// Start cleanup routine
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Daily cleanup
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := globalDB.CleanupOldData(30); err != nil { // Keep 30 days
					log.Printf("Failed to cleanup old data: %v", err)
				}
			}
		}
	}()

	return nil
}

// GetDB returns the global database instance
func GetDB() *DB {
	return globalDB
}
