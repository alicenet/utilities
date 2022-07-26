package service

import (
	"net"
	"path"
	"syscall"
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	_, s := NewServer()
	info := s.GetServiceInfo()

	if _, ok := info["grpc.reflection.v1alpha.ServerReflection"]; !ok {
		t.Error("reflection not set up")
	}
}

func TestNewServerGracefulStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sock := path.Join(dir, "socket")
	_, srv := NewServer()

	lis, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}

	wait := make(chan struct{})

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Error(err)
		}
		wait <- struct{}{}
	}()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Error(err)
	}

	<-wait
}

func TestMultipleServers(t *testing.T) {
	t.Parallel()
	t.Skip("should test this")
}
