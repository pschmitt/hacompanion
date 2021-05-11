package main

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type OnlineCheck struct {
	mode   string
	target string
	client http.Client
}

func NewOnlineCheck(m Meta) *OnlineCheck {
	o := OnlineCheck{
		mode: "ping",
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	if mode := m.GetString("mode"); mode != "" {
		o.mode = mode
	}
	if host := m.GetString("target"); host != "" {
		o.target = host
	}
	return &o
}

func (o OnlineCheck) run(ctx context.Context) (*payload, error) {
	if o.target == "" {
		return nil, fmt.Errorf("online check requires target to be specified")
	}
	switch o.mode {
	case "http":
		return o.checkHTTP(ctx)
	case "ping":
		return o.checkPing(ctx)
	default:
		return nil, fmt.Errorf("unknown mode for online check: %s", o.mode)
	}
}

func (o OnlineCheck) checkHTTP(ctx context.Context) (*payload, error) {
	p := NewPayload()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HomeAssistant-Companion/Online-Check")

	resp, err := o.client.Do(req)
	if err != nil {
		p.State = false
		p.Attributes["err"] = err.Error()
		return p, nil
	}

	p.State = true
	p.Attributes["status"] = resp.Status

	return p, nil
}

func (o OnlineCheck) checkPing(ctx context.Context) (*payload, error) {
	p := NewPayload()
	cmd := exec.CommandContext(ctx, "ping", "-c 2", "-w 4", o.target)
	err := cmd.Run()
	if err != nil {
		p.State = false
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			p.Attributes["err"] = fmt.Sprintf("could not reach %s", o.target)
		} else {
			p.Attributes["err"] = fmt.Sprintf("failed to execute ping: %s", err)
		}
		return p, nil
	}

	p.State = true
	return p, nil
}
