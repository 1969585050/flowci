package store

import "fmt"

func GetSettings() (map[string]string, error) {
	if DB == nil {
		return map[string]string{}, fmt.Errorf("database not initialized")
	}
	rows, err := DB.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings[key] = value
	}
	return settings, rows.Err()
}

func SaveSettings(key, value string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := DB.Exec(`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?`, key, value, value)
	if err != nil {
		return fmt.Errorf("save setting: %w", err)
	}
	return nil
}
