package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omerkaya1/grpc-file-stream/internal"
	"github.com/omerkaya1/grpc-file-stream/internal/domain"
	g "github.com/omerkaya1/grpc-file-stream/internal/grpc"
	"github.com/omerkaya1/grpc-file-stream/internal/grpc/api"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	// RootCmd is the root command for the project
	RootCmd = &cobra.Command{
		Use:     "file-streamer",
		Short:   "file-streamer is a small programme for file streaming through GRPC for educational purposes",
		Example: "file-streamer client -h",
	}
	clientCmd = &cobra.Command{
		Use:     "client",
		Short:   "invokes ",
		Example: "file-streamer client -c /path/to/config -f /path/to/file",
		Run:     client,
	}
	serverCmd = &cobra.Command{
		Use:     "server",
		Short:   "starts server",
		Example: "file-streamer server -c /path/to/config",
		RunE:    server,
	}
	cfg, file string
	timeout   int64
)

func init() {
	RootCmd.AddCommand(clientCmd, serverCmd)
	serverCmd.Flags().StringVarP(&cfg, "config", "c", "", "-c /path/to/config/file")
	clientCmd.Flags().StringVarP(&cfg, "config", "c", "", "-c /path/to/config/file")
	clientCmd.Flags().StringVarP(&file, "file", "f", "", "-f /path/to/file")
	clientCmd.Flags().Int64VarP(&timeout, "timeout", "t", 10, "-t 60")
}

func client(cmd *cobra.Command, args []string) {
	// Init configuration
	if cfg == "" || file == "" {
		log.Fatal("required arguments were not specified")
	}
	config, err := internal.InitConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// Validate it
	if config.Validate() {
		log.Fatal("configuration file did not pass validation")
	}
	conn, err := grpc.Dial(config.Host+":"+config.Port, grpc.WithInsecure())
	fsc := api.NewFileStreamerServiceClient(conn)

	result := make(chan error)
	stop := prepStopChan()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	go upload(ctx, file, fsc, result)

	select {
	case <-stop:
		log.Print("manual interruption; canceling context")
		cancel()
	case <-ctx.Done():
		log.Printf("context done: %s", ctx.Err())
		cancel()
	case err := <-result:
		if err != nil {
			log.Printf("we received an error from errChan: %s", err)
		}
		return
	}
	return
}

func server(cmd *cobra.Command, args []string) error {
	// Init configuration
	if cfg == "" {
		return errors.New("required arguments were not specified")
	}
	config, err := internal.InitConfig(cfg)
	if err != nil {
		return err
	}
	l := log.New(os.Stdout, "file-streamer: ", 0)
	// Store
	store, err := domain.NewStore(config.Destination, l)
	if err != nil {
		return err
	}
	// Signal channel
	stop := prepStopChan()

	ctx, cancel := context.WithCancel(context.Background())
	go func(cancel context.CancelFunc) {
		for s := range stop {
			l.Printf("received signal: %s", s)
			cancel()
		}
	}(cancel)

	return g.NewServer(store, l, config).Run(ctx)
}

func upload(ctx context.Context, path string, cc api.FileStreamerServiceClient, errChan chan error) {
	// Open the file
	f, err := os.Open(path)
	if err != nil {
		errChan <- err
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		errChan <- err
		return
	}
	// Prepare the upload
	c, err := cc.StoreFile(ctx)
	if err != nil {
		errChan <- err
		return
	}
	defer c.CloseSend()
	// At most 1 Kb at a time
	// buf := make([]byte, bytes.MinRead * 2)
	buf := make([]byte, bytes.MinRead*4)
	var written int64

WRITE:
	for {
		select {
		case <-ctx.Done():
			log.Printf("context was canceled; terminating request: written %d out of %d", written, info.Size())
			errChan <- ctx.Err()
			break WRITE
		default:
			// Fill the buffer, repeat
			n, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Printf("EOF reached: written %d out of %d", written, info.Size())
				} else {
					errChan <- err
				}
				break WRITE
			}

			written += int64(n)
			if err := c.Send(&api.Request{
				ResourceName: f.Name(),
				ReadOffset:   written,
				ReadLimit:    int64(n),
				Content:      buf[:n],
			}); err != nil {
				errChan <- err
				return
			}
		}
	}
	status, err := c.CloseAndRecv()
	if err != nil {
		log.Printf("exited [code] %d [message] %s [error] %s", status.Code, status.Message, err)
	}
	if status.Code != 0 {
		log.Printf("exited [code] %d [message] %s", status.Code, status.Message)
	}
	errChan <- err
	return
}

func prepStopChan() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	return stop
}
