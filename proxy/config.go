package proxy

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Config struct {
	mu sync.RWMutex

	Address          string `json:"address"`
	Nick             string `json:"nick"`
	DeviceModel      string `json:"deviceModel"`
	InputMode        int    `json:"inputMode"`
	DInputMode       int    `json:"defaultInputMode"`
	UIProfile        int    `json:"uiProfile"`
	DeviceOS         int    `json:"deviceOS"`
	BypassRPDownload bool   `json:"bypassRPDownload"`
	GameVersion      string `json:"gameVersion"`
	CurrentClientID  int64  `json:"-"`
	SkinData         string `json:"-"`
	SkinID           string `json:"-"`
	UUID             string `json:"-"`
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		cfg := &Config{
			Address:          "play.example.com:19132",
			DeviceModel:      "XIAOMI MT84373Y48J",
			InputMode:        2,
			BypassRPDownload: false,
			GameVersion:      "",
		}
		out, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(path, out, 0644)
		log.Println("[mefproxy] config.json создан. Укажи адрес сервера и перезапусти.")
		os.Exit(0)
	}
	cfg := &Config{}
	if err = json.Unmarshal(data, cfg); err != nil {
		log.Fatalln("[mefproxy] Ошибка чтения config.json:", err)
	}
	if cfg.Address == "" || cfg.Address == "play.example.com:19132" {
		log.Fatalln("[mefproxy] Укажи корректный адрес сервера в config.json")
	}
	if cfg.GameVersion != "" {
		log.Printf("[mefproxy] Спуф версии: %s", cfg.GameVersion)
	}
	return cfg
}

func (c *Config) Get() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c
}

func (c *Config) SetAddress(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Address = addr
}

func (c *Config) SetNick(nick string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Nick = nick
}

func (c *Config) SetBypassRP(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BypassRPDownload = v
}

func (c *Config) SetGameVersion(ver string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GameVersion = ver
}
