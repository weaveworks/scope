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

	TriggeringUser   string   `json:"triggering_user"`
	IntendedServices []string `json:"intended_services"`
}

// Config for the deployment system for a user.
type Config struct {
	RepoURL        string `json:"repo_url" yaml:"repo_url"`
	RepoBranch     string `json:"repo_branch" yaml:"repo_branch"`
	RepoPath       string `json:"repo_path" yaml:"repo_path"`
	RepoKey        string `json:"repo_key" yaml:"repo_key"`
	KubeconfigPath string `json:"kubeconfig_path" yaml:"kubeconfig_path"`
	AutoApply      bool   `json:"auto_apply" yaml:"auto_apply"`

	Notifications []NotificationConfig `json:"notifications" yaml:"notifications"`

	// Globs of files not to change, relative to the route of the repo
	ConfigFileBlackList []string `json:"config_file_black_list" yaml:"config_file_black_list"`

	CommitMessageTemplate string `json:"commit_message_template" yaml:"commit_message_template"` // See https://golang.org/pkg/text/template/
}

// NotificationConfig describes how to send notifications
type NotificationConfig struct {
	SlackWebhookURL      string `json:"slack_webhook_url" yaml:"slack_webhook_url"`
	SlackUsername        string `json:"slack_username" yaml:"slack_username"`
	MessageTemplate      string `json:"message_template" yaml:"message_template"`
	ApplyMessageTemplate string `json:"apply_message_template" yaml:"apply_message_template"`
}
