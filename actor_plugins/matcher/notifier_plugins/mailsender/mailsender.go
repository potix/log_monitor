package main

import (
    "log"
    "strings"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actor_plugins/matcher/notifierplugger"
    "github.com/potix/log_monitor/actor_plugins/matcher/notifier_plugins/mailsender/utility"
    "github.com/potix/log_monitor/actor_plugins/matcher/notifier_plugins/mailsender/configurator"
)

const (
	defaultSubjectFormat string = "${LABEL} - ${FILENAME}"
)

// MailSender is MailSender
type MailSender struct {
        callers string
        config *configurator.Config
        smtpClient *utility.SMTPClient
}

// Notify is notify
func (m *MailSender) Notify(msg []byte, fileID string, fileName string, label string) {
log.Printf("notify")
	format := defaultSubjectFormat
	if m.config.SubjectFormat != "" {
	    format = m.config.SubjectFormat
	}
        r := strings.NewReplacer("${LABEL}", label, "${FILEID}", fileID, "${FILENAME}", fileName)
        subject := r.Replace(format)
        err := m.smtpClient.SendMail(subject, string(msg))
        if err != nil {
            log.Printf("can not send mail (%v, %v, %v, %v): err", m.config.From, m.config.To, m.config.HostPort, subject)
        }
}

// NewMailSender is create new mail sender
func NewMailSender(callers string, configFile string) (notifierplugger.NotifierPlugin, error) {
    log.Printf("configFile = %v", configFile)
    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }
    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }
    log.Printf("config = %v", config)
    newCallers := callers + ".mailsender"
    smtpClient := utility.NewSMTPClient(config.HostPort, config.Username,
        config.Password, utility.GetSMTPAuthType(config.AuthType),
        config.UseTLS, config.UseStartTLS, config.From, config.To)
    return &MailSender {
        callers: newCallers,
        config: config,
        smtpClient: smtpClient,
    }, nil
}

// GetNotifierPluginInfo is GetNotifierPluginInfo
func GetNotifierPluginInfo() (string, notifierplugger.NotifierPluginNewFunc) {
    return "mailsender", NewMailSender
}
