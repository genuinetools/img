package fsutils

import (
	"os"

	"github.com/pkg/errors"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/net/context"
)

type dynamicWalker struct {
	walkChan chan *currentPath
	err      error
	closeCh  chan struct{}
}

func newDynamicWalker() *dynamicWalker {
	return &dynamicWalker{
		walkChan: make(chan *currentPath, 128),
		closeCh:  make(chan struct{}),
	}
}

func (w *dynamicWalker) update(p *currentPath) error {
	select {
	case <-w.closeCh:
		return errors.Wrap(w.err, "walker is closed")
	default:
	}
	if p == nil {
		close(w.walkChan)
		return nil
	}
	select {
	case w.walkChan <- p:
		return nil
	case <-w.closeCh:
		return errors.Wrap(w.err, "walker is closed")
	}
}

func (w *dynamicWalker) fill(ctx context.Context, pathC chan<- *currentPath) error {
	for {
		select {
		case p, ok := <-w.walkChan:
			if !ok {
				return nil
			}
			pathC <- p
		case <-ctx.Done():
			w.err = ctx.Err()
			close(w.closeCh)
			return ctx.Err()
		}
	}
}

func getWalkerFn(root string) walkerFn {
	return func(ctx context.Context, pathC chan<- *currentPath) error {
		return fsutil.Walk(ctx, root, nil, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			p := &currentPath{
				path: path,
				f:    f,
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case pathC <- p:
				return nil
			}
		})
	}
}
