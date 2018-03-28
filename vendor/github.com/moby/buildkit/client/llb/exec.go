package llb

import (
	_ "crypto/sha256"
	"sort"

	"github.com/moby/buildkit/solver/pb"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type Meta struct {
	Args []string
	Env  EnvList
	Cwd  string
	User string
}

func NewExecOp(root Output, meta Meta, readOnly bool, md OpMetadata) *ExecOp {
	e := &ExecOp{meta: meta, cachedOpMetadata: md}
	rootMount := &mount{
		target:   pb.RootMount,
		source:   root,
		readonly: readOnly,
	}
	e.mounts = append(e.mounts, rootMount)
	if readOnly {
		e.root = root
	} else {
		e.root = &output{vertex: e, getIndex: e.getMountIndexFn(rootMount)}
	}
	rootMount.output = e.root

	return e
}

type mount struct {
	target   string
	readonly bool
	source   Output
	output   Output
	selector string
	// hasOutput bool
}

type ExecOp struct {
	root             Output
	mounts           []*mount
	meta             Meta
	cachedPBDigest   digest.Digest
	cachedPB         []byte
	cachedOpMetadata OpMetadata
	isValidated      bool
}

func (e *ExecOp) AddMount(target string, source Output, opt ...MountOption) Output {
	m := &mount{
		target: target,
		source: source,
	}
	for _, o := range opt {
		o(m)
	}
	e.mounts = append(e.mounts, m)
	if m.readonly {
		m.output = source
	} else {
		m.output = &output{vertex: e, getIndex: e.getMountIndexFn(m)}
	}
	e.cachedPB = nil
	e.isValidated = false
	return m.output
}

func (e *ExecOp) GetMount(target string) Output {
	for _, m := range e.mounts {
		if m.target == target {
			return m.output
		}
	}
	return nil
}

func (e *ExecOp) Validate() error {
	if e.isValidated {
		return nil
	}
	if len(e.meta.Args) == 0 {
		return errors.Errorf("arguments are required")
	}
	if e.meta.Cwd == "" {
		return errors.Errorf("working directory is required")
	}
	for _, m := range e.mounts {
		if m.source != nil {
			if err := m.source.Vertex().Validate(); err != nil {
				return err
			}
		}
	}
	e.isValidated = true
	return nil
}

func (e *ExecOp) Marshal() (digest.Digest, []byte, *OpMetadata, error) {
	if e.cachedPB != nil {
		return e.cachedPBDigest, e.cachedPB, &e.cachedOpMetadata, nil
	}
	if err := e.Validate(); err != nil {
		return "", nil, nil, err
	}
	// make sure mounts are sorted
	sort.Slice(e.mounts, func(i, j int) bool {
		return e.mounts[i].target < e.mounts[j].target
	})

	peo := &pb.ExecOp{
		Meta: &pb.Meta{
			Args: e.meta.Args,
			Env:  e.meta.Env.ToArray(),
			Cwd:  e.meta.Cwd,
			User: e.meta.User,
		},
	}

	pop := &pb.Op{
		Op: &pb.Op_Exec{
			Exec: peo,
		},
	}

	outIndex := 0
	for _, m := range e.mounts {
		inputIndex := pb.InputIndex(len(pop.Inputs))
		if m.source != nil {
			inp, err := m.source.ToInput()
			if err != nil {
				return "", nil, nil, err
			}

			newInput := true

			for i, inp2 := range pop.Inputs {
				if *inp == *inp2 {
					inputIndex = pb.InputIndex(i)
					newInput = false
					break
				}
			}

			if newInput {
				pop.Inputs = append(pop.Inputs, inp)
			}
		} else {
			inputIndex = pb.Empty
		}

		outputIndex := pb.OutputIndex(-1)
		if !m.readonly {
			outputIndex = pb.OutputIndex(outIndex)
			outIndex++
		}

		pm := &pb.Mount{
			Input:    inputIndex,
			Dest:     m.target,
			Readonly: m.readonly,
			Output:   outputIndex,
			Selector: m.selector,
		}
		peo.Mounts = append(peo.Mounts, pm)
	}

	dt, err := pop.Marshal()
	if err != nil {
		return "", nil, nil, err
	}
	e.cachedPBDigest = digest.FromBytes(dt)
	e.cachedPB = dt
	return e.cachedPBDigest, dt, &e.cachedOpMetadata, nil
}

func (e *ExecOp) Output() Output {
	return e.root
}

func (e *ExecOp) Inputs() (inputs []Output) {
	mm := map[Output]struct{}{}
	for _, m := range e.mounts {
		if m.source != nil {
			mm[m.source] = struct{}{}
		}
	}
	for o := range mm {
		inputs = append(inputs, o)
	}
	return
}

func (e *ExecOp) getMountIndexFn(m *mount) func() (pb.OutputIndex, error) {
	return func() (pb.OutputIndex, error) {
		// make sure mounts are sorted
		sort.Slice(e.mounts, func(i, j int) bool {
			return e.mounts[i].target < e.mounts[j].target
		})

		i := 0
		for _, m2 := range e.mounts {
			if m2.readonly {
				continue
			}
			if m == m2 {
				return pb.OutputIndex(i), nil
			}
			i++
		}
		return pb.OutputIndex(0), errors.Errorf("invalid mount")
	}
}

type ExecState struct {
	State
	exec *ExecOp
}

func (e ExecState) AddMount(target string, source State, opt ...MountOption) State {
	return source.WithOutput(e.exec.AddMount(target, source.Output(), opt...))
}

func (e ExecState) GetMount(target string) State {
	return NewState(e.exec.GetMount(target))
}

func (e ExecState) Root() State {
	return e.State
}

type MountOption func(*mount)

func Readonly(m *mount) {
	m.readonly = true
}

func SourcePath(src string) MountOption {
	return func(m *mount) {
		m.selector = src
	}
}

type RunOption interface {
	SetRunOption(es *ExecInfo)
}

type runOptionFunc func(*ExecInfo)

func (fn runOptionFunc) SetRunOption(ei *ExecInfo) {
	fn(ei)
}

func Shlex(str string) RunOption {
	return Shlexf(str)
}
func Shlexf(str string, v ...interface{}) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = shlexf(str, v...)(ei.State)
	})
}

func Args(a []string) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = args(a...)(ei.State)
	})
}

func AddEnv(key, value string) RunOption {
	return AddEnvf(key, value)
}

func AddEnvf(key, value string, v ...interface{}) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = ei.State.AddEnvf(key, value, v...)
	})
}

func User(str string) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = ei.State.User(str)
	})
}

func Dir(str string) RunOption {
	return Dirf(str)
}
func Dirf(str string, v ...interface{}) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = ei.State.Dirf(str, v...)
	})
}

func Reset(s State) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = ei.State.Reset(s)
	})
}

func With(so ...StateOption) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.State = ei.State.With(so...)
	})
}

func AddMount(dest string, mountState State, opts ...MountOption) RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.Mounts = append(ei.Mounts, MountInfo{dest, mountState.Output(), opts})
	})
}

func ReadonlyRootFS() RunOption {
	return runOptionFunc(func(ei *ExecInfo) {
		ei.ReadonlyRootFS = true
	})
}

type ExecInfo struct {
	opMetaWrapper
	State          State
	Mounts         []MountInfo
	ReadonlyRootFS bool
}

type MountInfo struct {
	Target string
	Source Output
	Opts   []MountOption
}
