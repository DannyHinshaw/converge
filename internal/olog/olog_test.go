package olog_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannyhinshaw/converge/internal/olog"
)

func TestLogger_Debug(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelDebug, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Debug("debug message")

	expected := fmt.Sprintf("[%s] [TestLogger]: debug message\n", olog.LevelDebug)
	a.Contains(buf.String(), expected)
}

func TestLogger_Debugf(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelDebug, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Debugf("debug message %d", 1)

	expected := fmt.Sprintf("[%s] [TestLogger]: debug message 1\n", olog.LevelDebug)
	a.Contains(buf.String(), expected)
}

func TestLogger_Info(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelInfo, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Info("info message")

	expected := fmt.Sprintf("[%s] [TestLogger]: info message\n", olog.LevelInfo)
	a.Contains(buf.String(), expected)
}

func TestLogger_Infof(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelInfo, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Infof("info message %d", 1)

	expected := fmt.Sprintf("[%s] [TestLogger]: info message 1\n", olog.LevelInfo)
	a.Contains(buf.String(), expected)
}

func TestLogger_Error(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelError, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Error("error message")

	expected := fmt.Sprintf("[%s] [TestLogger]: error message\n", olog.LevelError)
	a.Contains(buf.String(), expected)
}

func TestLogger_Errorf(t *testing.T) {
	a := assert.New(t)

	var buf bytes.Buffer
	logger := olog.NewLogger(olog.LevelError, olog.WithWriter(&buf)).
		WithName("TestLogger")

	logger.Errorf("error message %d", 1)

	expected := fmt.Sprintf("[%s] [TestLogger]: error message 1\n", olog.LevelError)
	a.Contains(buf.String(), expected)
}
