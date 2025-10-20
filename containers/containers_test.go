package containers

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UseCaseFunc func(name string) string

func UseCaseFuncImpl(name string) string {
	return name
}

type UseCase interface {
	Execute() bool
}

type UseCaseA struct {
	A int
}

func (u *UseCaseA) Execute() bool {
	return true
}

func TestContainer_Register(t *testing.T) {
	expect := assert.New(t)
	c := New()

	c.Register(func() UseCaseA {
		return UseCaseA{A: 1}
	})

	var x1 UseCaseA
	err := c.Resolve(&x1)
	expect.IsType(UseCaseA{}, x1)
	expect.NoError(err)

	var x2 UseCaseA
	err = c.Resolve(&x2)
	expect.IsType(UseCaseA{}, x2)
	expect.NoError(err)

	c.Register(func() UseCaseFunc {
		return UseCaseFuncImpl
	})
	var uc UseCaseFunc
	err = c.Resolve(&uc)
	expect.NoError(err)
	expect.Equal("n1", uc("n1"))

	expect.NotSame(x1, x2)
}

func TestContainer_Register_Pointer(t *testing.T) {
	expect := assert.New(t)
	c := New()

	c.Register(func() *UseCaseA {
		return &UseCaseA{A: 1}
	})

	var x1 *UseCaseA
	err := c.Resolve(&x1)
	expect.IsType(&UseCaseA{}, x1)
	expect.NoError(err)

	var x2 *UseCaseA
	err = c.Resolve(&x2)
	expect.IsType(&UseCaseA{}, x2)
	expect.NoError(err)

	expect.NotSame(x1, x2)
}

func TestContainer_Singleton(t *testing.T) {
	expect := assert.New(t)
	c := New()

	c.Singleton(func() *UseCaseA {
		return &UseCaseA{A: 1}
	})

	var x1 *UseCaseA
	err := c.Resolve(&x1)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x1)

	var x2 *UseCaseA
	err = c.Resolve(&x2)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x2)

	c.Singleton(func() UseCaseFunc {
		return UseCaseFuncImpl
	})
	var uc UseCaseFunc
	err = c.Resolve(&uc)
	expect.NoError(err)
	expect.Equal("n1", uc("n1"))

	expect.Same(x1, x2)
}

func TestContainer_Singleton_Interface(t *testing.T) {
	expect := assert.New(t)
	c := New()

	c.Singleton(func() UseCase {
		return &UseCaseA{A: 1}
	})

	var x1 UseCase
	err := c.Resolve(&x1)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x1)

	var x2 UseCase
	err = c.Resolve(&x2)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x2)

	expect.Same(x1, x2)
}

func TestContainer_Value(t *testing.T) {
	expect := assert.New(t)

	c := New()

	x1 := &UseCaseA{A: 1}
	c.Value(x1)

	var d1 *UseCaseA
	err := c.Resolve(&d1)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, d1)
	expect.Same(x1, d1)
}

func TestProvider_Not_Found(t *testing.T) {
	expect := assert.New(t)

	c := New()
	var x1 UseCase
	err := c.Resolve(&x1)
	expect.Error(err)
	expect.Equal("no resolver found for: containers.UseCase", err.Error())
}

func TestContainer_Singleton_Race_Resolve(t *testing.T) {
	expect := assert.New(t)
	c := New()

	c.Singleton(func() *UseCaseA {
		return &UseCaseA{A: 1}
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	var x1 *UseCaseA
	var err1 error
	go func() {
		err1 = c.Resolve(&x1)
		wg.Done()
	}()

	var x2 *UseCaseA
	var err2 error
	go func() {
		err2 = c.Resolve(&x2)
		wg.Done()
	}()

	wg.Wait()

	expect.NoError(err1)
	expect.IsType(&UseCaseA{}, x1)

	expect.NoError(err2)
	expect.IsType(&UseCaseA{}, x2)

	expect.Same(x1, x2)
}

type Database interface {
	Read() int
}

type DatabaseImpl struct {
	A int
}

func (d *DatabaseImpl) Read() int {
	return d.A
}

type Repo struct {
	Db1 Database `inject:"" json:"db1"`
	Db2 Database `inject:"d1" json:"db2"`
}

func TestContainer_Inject(t *testing.T) {
	expect := assert.New(t)

	c := New()

	c.Singleton(func() Database {
		return &DatabaseImpl{A: 1}
	}).NamedSingleton("d1", func() Database {
		return &DatabaseImpl{A: 2}
	})

	repo := &Repo{}
	err := c.Inject(repo)

	expect.NoError(err)
	expect.NotNil(repo.Db1)
	expect.NotNil(repo.Db2)
	expect.Equal(1, repo.Db1.Read())
	expect.Equal(2, repo.Db2.Read())
}
