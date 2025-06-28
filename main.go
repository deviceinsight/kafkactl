// Copyright © 2018 Dirk Wilden <dirk.wilden@device-insight.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deviceinsight/kafkactl/v5/cmd"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := runWithShutdownTimeout(ctx, 30*time.Second); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			output.Warnf("Shutdown timeout exceeded, forced exit")
			os.Exit(1)
		}
		output.Warnf("%v", err)
		os.Exit(1)
	}
}

func runWithShutdownTimeout(ctx context.Context, shutdownTimeout time.Duration) error {
	errChan := make(chan error, 1)
	ioStreams := output.DefaultIOStreams()

	rootCmd := cmd.NewKafkactlCommand(ioStreams)

	go func() {
		errChan <- rootCmd.ExecuteContext(ctx)
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		output.Debugf("Shutdown signal received, allowing %v for graceful shutdown...", shutdownTimeout)

		select {
		case err := <-errChan:
			return err
		case <-time.Tick(shutdownTimeout):
			return errors.Join(context.DeadlineExceeded, ctx.Err())
		}
	}
}
