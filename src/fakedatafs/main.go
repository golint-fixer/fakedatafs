package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/jessevdk/go-flags"
)

var (
	version    = "compiled manually"
	compiledAt = "unknown"
)

// Options are global settings.
type Options struct {
	Seed    int  `long:"seed"                  description:"initial random seed"`
	Version bool `long:"version" short:"v"     description:"print version number"`

	mountpoint string
}

var opts = Options{}
var parser = flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)

var exitRequested = make(chan struct{})

func init() {
	parser.Usage = "mountpoint"
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	go cleanupHandler(c)
}

func cleanupHandler(c <-chan os.Signal) {
	once := &sync.Once{}

	for range c {
		once.Do(func() {
			fmt.Println("Interrupt received, cleaning up")
			close(exitRequested)
		})
	}
}

func mount(opts Options) error {
	conn, err := fuse.Mount(
		opts.mountpoint,
		fuse.ReadOnly(),
		fuse.FSName("fakedatafs"),
	)
	if err != nil {
		return err
	}

	root := fs.Tree{}

	// root.Add("snapshots", fuse.NewSnapshotsDir(repo, cmd.Root))

	// cmd.global.Printf("Now serving %s at %s\n", repo.Backend().Location(), mountpoint)
	// cmd.global.Printf("Don't forget to umount after quitting!\n")

	// AddCleanupHandler(func() error {
	// 	return fuse.Unmount(mountpoint)
	// })

	// cmd.ready <- struct{}{}

	err = fs.Serve(conn, &root)
	if err != nil {
		return err
	}

	<-conn.Ready
	err = conn.MountError
	if err != nil {
		fmt.Fprintf(os.Stderr, "mount failed: %v\n", err)
		return fuse.Unmount(opts.mountpoint)
	}

	fmt.Printf("successfully mounted fakedatafs at %v\n", opts.mountpoint)

	<-exitRequested

	fmt.Printf("umounting...\n")
	return fuse.Unmount(opts.mountpoint)
}

func main() {
	args, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("version %v, compiled at %v using %v\n", version, compiledAt, runtime.Version())
		return
	}

	if len(args) == 0 {
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	opts.mountpoint = args[0]
	err = mount(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
