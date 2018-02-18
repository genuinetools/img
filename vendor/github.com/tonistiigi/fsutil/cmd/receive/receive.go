package main

import (
	"flag"
	"os"

	"github.com/tonistiigi/fsutil"
	"github.com/tonistiigi/fsutil/util"
	"golang.org/x/net/context"
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		panic("dest path not set")
	}

	ctx := context.Background()
	s := util.NewProtoStream(ctx, os.Stdin, os.Stdout)

	if err := fsutil.Receive(ctx, s, flag.Args()[0], fsutil.ReceiveOpt{}); err != nil {
		panic(err)
	}
}
