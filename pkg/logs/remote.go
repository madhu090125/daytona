// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type remoteLoggerFactory struct {
	loggerFactoryImpl *loggerFactory
	serverUrl         string
	apiKey            string
}

func NewRemoteLoggerFactory(targetLogsDir *string, buildLogsDir *string, serverUrl string, apiKey string) ILoggerFactory {
	loggerFactoryImpl := &remoteLoggerFactory{
		loggerFactoryImpl: &loggerFactory{},
		serverUrl:         serverUrl,
		apiKey:            apiKey,
	}

	if targetLogsDir != nil {
		loggerFactoryImpl.loggerFactoryImpl.targetLogsDir = *targetLogsDir
	}

	if buildLogsDir != nil {
		loggerFactoryImpl.loggerFactoryImpl.buildLogsDir = *buildLogsDir
	}

	return loggerFactoryImpl
}

func (l *remoteLoggerFactory) CreateWorkspaceLogger(workspaceId, workspaceName string, source LogSource) (Logger, error) {
	logger := logrus.New()

	conn, _, err := util.GetWebsocketConn(context.Background(), "/log/workspace/"+workspaceId, l.serverUrl, l.apiKey, nil)
	if err != nil {
		return nil, err
	}

	return &remoteLogger{
		Logger: &workspaceLogger{
			workspaceId:   workspaceId,
			logsDir:       l.loggerFactoryImpl.targetLogsDir,
			workspaceName: workspaceName,
			logger:        logger,
			source:        source,
		},
		conn: conn,
	}, nil
}

func (l *remoteLoggerFactory) CreateTargetLogger(targetId, targetName string, source LogSource) (Logger, error) {
	logger := logrus.New()

	conn, _, err := util.GetWebsocketConn(context.Background(), "/log/target/"+targetId, l.serverUrl, l.apiKey, nil)
	if err != nil {
		return nil, err
	}

	return &remoteLogger{
		Logger: &targetLogger{
			targetId:   targetId,
			targetName: targetName,
			logsDir:    l.loggerFactoryImpl.targetLogsDir,
			logger:     logger,
			source:     source,
		},
		conn: conn,
	}, nil
}

func (l *remoteLoggerFactory) CreateBuildLogger(buildId string, source LogSource) (Logger, error) {
	logger := logrus.New()

	conn, _, err := util.GetWebsocketConn(context.Background(), "/log/build/"+buildId, l.serverUrl, l.apiKey, nil)
	if err != nil {
		return nil, err
	}

	return &remoteLogger{
		Logger: &buildLogger{
			logsDir: l.loggerFactoryImpl.buildLogsDir,
			buildId: buildId,
			logger:  logger,
			source:  source,
		},
		conn: conn,
	}, nil
}

func (l *remoteLoggerFactory) CreateWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (l *remoteLoggerFactory) CreateTargetLogReader(targetId string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (l *remoteLoggerFactory) CreateBuildLogReader(buildId string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

type remoteLogger struct {
	Logger
	conn *websocket.Conn
}

func (b *remoteLogger) Write(p []byte) (n int, err error) {
	if b.conn != nil {
		err = b.conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return len(p), err
		}
	}

	return b.Write(p)
}
