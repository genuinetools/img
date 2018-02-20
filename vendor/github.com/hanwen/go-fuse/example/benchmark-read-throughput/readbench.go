// Copyright 2017 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// readbench is a benchmark helper for measuring throughput on
// single-file reads out of a FUSE filesystem.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func gulp(fn string, bs int) (int, error) {
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var tot int
	buf := make([]byte, bs)
	for {
		n, _ := f.Read(buf[:])
		tot += n
		if n < len(buf) {
			break
		}
	}

	return tot, nil
}

func main() {
	bs := flag.Int("bs", 32, "blocksize in kb")
	mbLimit := flag.Int("limit", 1000, "amount of data to read in mb")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("readbench [-bs BLOCKSIZE -limit SIZE] file")
	}
	blocksize := *bs * 1024
	totMB := 0.0
	var totDT time.Duration

	for totMB < float64(*mbLimit) {
		t := time.Now()
		n, err := gulp(flag.Arg(0), blocksize)
		if err != nil {
			log.Fatal(err)
		}
		dt := time.Now().Sub(t)
		mb := float64(n) / (1 << 20)

		totMB += mb
		totDT += dt
	}
	fmt.Printf("block size %d kb: %.1f MB in %v: %.2f MBs/s\n", *bs, totMB, totDT, totMB/float64(totDT)*1e9)
}
