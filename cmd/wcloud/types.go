package main

import (
	"time"
)

// Deployment describes a deployment
type Deployment struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ImageName string    `json:"image_name"`
	Version   string    `json:"version"`
	Priority  int       `json:"priority"`
	State     string    `json:"status"`
	LogKey    string    `json:"-"`
}

// Config for the deployment system for a user.
type Config struct {
	RepoURL        string `json:"repo_url" yaml:"repo_url"`
	RepoPath       string `json:"repo_path" yaml:"repo_path"`
	RepoKey        string `json:"repo_key" yaml:"repo_key"`
	KubeconfigPath string `json:"kubeconfig_path" yaml:"kubeconfig_path"`

	Notifications []NotificationConfig `json:"notifications" yaml:"notifications"`

	// Globs of files not to change, relative to the route of the repo
	ConfigFileBlackList []string `json:"config_file_black_list" yaml:"config_file_black_list"`
}

// NotificationConfig describes how to send notifications
type NotificationConfig struct {
	SlackWebhookURL string `json:"slack_webhook_url" yaml:"slack_webhook_url"`
	SlackUsername   string `json:"slack_username" yaml:"slack_username"`
}
