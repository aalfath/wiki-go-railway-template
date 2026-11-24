// wiki-go-init prepares Wiki-Go's data/config.yaml for Railway before the app starts.
//
//   - First boot (no config.yaml): writes a config whose only user is the admin from
//     WIKI_ADMIN_USERNAME / WIKI_ADMIN_PASSWORD (bcrypt), so upstream's default
//     admin/admin account is never created.
//   - Railway mounts volumes owned by root, so when started as root it gives the
//     data directory to the image's appuser (uid 1000), then drops to that user.
//   - Every boot: listens on $PORT on all interfaces, trusts Railway's proxy for
//     client IPs (login ban), and replaces the admin password if it is still the
//     upstream default "admin" or if WIKI_RESET_ADMIN_PASSWORD=true.
//
// It then execs Wiki-Go (PID 1 stays the app, so a crash exits non-zero).
// Exits non-zero on any error so Railway restarts the deploy instead of running
// an unsafe wiki.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

const cost = 14 // upstream's default security.passwordstrength

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "wiki-go-init: "+format+"\n", a...)
	os.Exit(1)
}

func section(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	v := map[string]any{}
	m[key] = v
	return v
}

const appUID, appGID = 1000, 1000

// dropPrivileges hands the data directory to appuser and switches to it.
func dropPrivileges(dataDir string) {
	if os.Getuid() != 0 {
		return
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		fail("create %s: %v", dataDir, err)
	}
	err := filepath.Walk(dataDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok && st.Uid == appUID && st.Gid == appGID {
			return nil
		}
		return os.Lchown(p, appUID, appGID)
	})
	if err != nil {
		fail("chown %s: %v", dataDir, err)
	}
	if err := syscall.Setgroups([]int{appGID}); err != nil {
		fail("setgroups: %v", err)
	}
	if err := syscall.Setgid(appGID); err != nil {
		fail("setgid: %v", err)
	}
	if err := syscall.Setuid(appUID); err != nil {
		fail("setuid: %v", err)
	}
	os.Setenv("HOME", "/home/appuser")
	os.Setenv("USER", "appuser")
}

func main() {
	path := env("CONFIGFILE", "data/config.yaml")
	dropPrivileges(filepath.Dir(path))
	port, err := strconv.Atoi(env("PORT", "8080"))
	if err != nil || port < 1 || port > 65535 {
		fail("PORT must be a TCP port number, got %q", os.Getenv("PORT"))
	}
	user := env("WIKI_ADMIN_USERNAME", "admin")
	pass := os.Getenv("WIKI_ADMIN_PASSWORD")
	reset := strings.EqualFold(env("WIKI_RESET_ADMIN_PASSWORD", "false"), "true")
	var proxies []any
	for _, p := range strings.Split(env("WIKI_TRUSTED_PROXIES", ""), ",") {
		if p = strings.TrimSpace(p); p != "" {
			proxies = append(proxies, p)
		}
	}

	cfg := map[string]any{}
	data, err := os.ReadFile(path)
	first := errors.Is(err, os.ErrNotExist)
	switch {
	case first:
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fail("create %s: %v", filepath.Dir(path), err)
		}
	case err != nil:
		fail("read %s: %v", path, err)
	default:
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			fail("parse %s: %v", path, err)
		}
		if cfg == nil {
			cfg = map[string]any{}
		}
	}

	server := section(cfg, "server")
	server["host"] = "" // ":PORT" listens on IPv4 and IPv6 (Railway private network)
	server["port"] = port
	server["ssl"] = false // Railway terminates TLS at its edge
	server["trusted_proxies"] = proxies
	if first {
		server["allow_insecure_cookies"] = false
		wiki := section(cfg, "wiki")
		wiki["timezone"] = env("TZ", "UTC")
		if d := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); d != "" {
			wiki["owner"] = d
		}
		section(cfg, "security")["passwordstrength"] = cost
	}

	var users []any
	if u, ok := cfg["users"].([]any); ok {
		users = u
	}
	hasAdmin, idx := false, -1
	for i, raw := range users {
		u, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := u["username"].(string)
		role, _ := u["role"].(string)
		hash, _ := u["password"].(string)
		if role == "admin" {
			hasAdmin = true
		}
		if name == user {
			idx = i
		}
		// Never leave upstream's default admin/admin login in place.
		if role == "admin" && hash != "" && bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin")) == nil {
			if len(pass) < 12 {
				fail("user %q still has the default password and WIKI_ADMIN_PASSWORD is missing or shorter than 12 characters", name)
			}
			h, _ := bcrypt.GenerateFromPassword([]byte(pass), cost)
			u["password"] = string(h)
			fmt.Printf("wiki-go-init: replaced the default password of admin user %q\n", name)
		}
	}
	if first || !hasAdmin || reset {
		if len(pass) < 12 {
			fail("WIKI_ADMIN_PASSWORD is missing or shorter than 12 characters")
		}
		h, err := bcrypt.GenerateFromPassword([]byte(pass), cost)
		if err != nil {
			fail("hash password: %v", err)
		}
		if idx >= 0 {
			u := users[idx].(map[string]any)
			u["password"] = string(h)
			u["role"] = "admin"
		} else {
			users = append(users, map[string]any{"username": user, "password": string(h), "role": "admin"})
		}
		fmt.Printf("wiki-go-init: admin user %q set from WIKI_ADMIN_PASSWORD\n", user)
	}
	cfg["users"] = users

	out, err := yaml.Marshal(cfg)
	if err != nil {
		fail("encode config: %v", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		fail("write %s: %v", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		fail("replace %s: %v", path, err)
	}
	if first {
		fmt.Println("wiki-go-init: first boot, wrote", path)
	}

	bin := env("WIKI_GO_BIN", "/wiki/wiki-go")
	args := append([]string{bin}, os.Args[1:]...)
	if err := syscall.Exec(bin, args, os.Environ()); err != nil {
		fail("exec %s: %v", bin, err)
	}
}
