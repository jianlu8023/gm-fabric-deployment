package authz

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type authzLogger struct {
	zapLogger *zap.SugaredLogger
	enabled   bool
}

func (l *authzLogger) EnableLog(enabled bool) {
	l.enabled = enabled
}

func (l *authzLogger) IsEnabled() bool {
	return l.enabled
}

func (l *authzLogger) LogModel(model [][]string) {
	if !l.enabled {
		return
	}
	var str strings.Builder
	str.WriteString("Model: ")
	for _, v := range model {
		str.WriteString(fmt.Sprintf("%v\n", v))
	}
	l.zapLogger.Infof(str.String())
}

func (l *authzLogger) LogEnforce(matcher string, request []interface{}, result bool, explains [][]string) {
	if !l.enabled {
		return
	}

	var reqStr strings.Builder
	reqStr.WriteString("Request: ")
	for i, rval := range request {
		if i != len(request)-1 {
			reqStr.WriteString(fmt.Sprintf("%v, ", rval))
		} else {
			reqStr.WriteString(fmt.Sprintf("%v", rval))
		}
	}
	reqStr.WriteString(fmt.Sprintf(" ---> %t\n", result))

	reqStr.WriteString("Hit Policy: ")
	for i, pval := range explains {
		if i != len(explains)-1 {
			reqStr.WriteString(fmt.Sprintf("%v, ", pval))
		} else {
			reqStr.WriteString(fmt.Sprintf("%v \n", pval))
		}
	}
	l.zapLogger.Infof(reqStr.String())
}

func (l *authzLogger) LogRole(roles []string) {
	if !l.enabled {
		return
	}

	l.zapLogger.Infof("Roles: %v", strings.Join(roles, "\n"))
}

func (l *authzLogger) LogPolicy(policy map[string][][]string) {
	if !l.enabled {
		return
	}

	var str strings.Builder
	str.WriteString("Policy: ")
	for k, v := range policy {
		str.WriteString(fmt.Sprintf("%s : %v\n", k, v))
	}
	l.zapLogger.Infof(str.String())
}

func (l *authzLogger) LogError(err error, msg ...string) {
	if !l.enabled {
		return
	}
	l.zapLogger.Errorln(msg, err)
}

func newAuthzLogger(logger *zap.SugaredLogger, enabled bool) *authzLogger {
	named := logger.Named("casbin")
	return &authzLogger{
		zapLogger: named,
		enabled:   enabled,
	}
}
