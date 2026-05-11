package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// defaultDatabaseURL builds a Postgres URL from individual DATABASE_* env vars.
// When DATABASE_HOST is unset it probes a list of candidate hosts (gateway IP,
// Docker bridge, host.docker.internal, 127.0.0.1) and picks the first one that
// accepts a TCP connection on the database port.
func defaultDatabaseURL() string {
	user := getenvDefault("DATABASE_USER", "telemetry")
	pass := getenvDefault("DATABASE_PASSWORD", "telemetry")
	port := getenvDefault("DATABASE_PORT", "5433")
	db := getenvDefault("DATABASE_DB", "telemetry")
	host := defaultDBHost(port)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func getenvDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func defaultDBHost(port string) string {
	if h := strings.TrimSpace(os.Getenv("DATABASE_HOST")); h != "" {
		return h
	}
	candidates := candidateDBHosts()
	if h := firstHostAcceptingTCP(candidates, port, 800*time.Millisecond); h != "" {
		return h
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "127.0.0.1"
}

func candidateDBHosts() []string {
	var c []string
	if inDocker() && runtime.GOOS == "linux" {
		if gw, err := linuxDefaultGateway(); err == nil && gw != "" {
			c = append(c, gw)
		}
		c = append(c, "172.17.0.1", "172.18.0.1", "192.168.65.254", "host.docker.internal", "127.0.0.1")
		return dedupeHosts(c)
	}
	if inDocker() {
		return dedupeHosts([]string{"host.docker.internal", "172.17.0.1", "127.0.0.1"})
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return dedupeHosts([]string{"host.docker.internal", "127.0.0.1"})
	}
	return []string{"127.0.0.1"}
}

func dedupeHosts(hosts []string) []string {
	seen := make(map[string]struct{}, len(hosts))
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}
	return out
}

func firstHostAcceptingTCP(hosts []string, port string, perHost time.Duration) string {
	for _, h := range hosts {
		addr := net.JoinHostPort(h, port)
		c, err := net.DialTimeout("tcp", addr, perHost)
		if err == nil {
			_ = c.Close()
			return h
		}
	}
	return ""
}

func inDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func linuxDefaultGateway() (string, error) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		dest, gwHex := fields[1], fields[2]
		if dest != "00000000" || gwHex == "00000000" {
			continue
		}
		ip, err := gatewayHexToIP(gwHex)
		if err != nil {
			return "", err
		}
		if ip.IsUnspecified() {
			continue
		}
		return ip.String(), nil
	}
	return "", errors.New("no default gateway in /proc/net/route")
}

func gatewayHexToIP(hex string) (net.IP, error) {
	if len(hex) != 8 {
		return nil, fmt.Errorf("bad gateway field %q", hex)
	}
	oct := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, err
		}
		oct[3-i] = byte(b)
	}
	return net.IPv4(oct[0], oct[1], oct[2], oct[3]), nil
}
