package containers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal_Register(t *testing.T) {
	expect := assert.New(t)

	Register(func() UseCaseA {
		return UseCaseA{A: 1}
	})

	var x1 UseCaseA
	err := Resolve(&x1)
	expect.IsType(UseCaseA{}, x1)
	expect.NoError(err)

	var x2 UseCaseA
	err = Resolve(&x2)
	expect.IsType(UseCaseA{}, x2)
	expect.NoError(err)

	expect.NotSame(x1, x2)
}

func TestGlobal_Singleton(t *testing.T) {
	expect := assert.New(t)

	Singleton(func() *UseCaseA {
		return &UseCaseA{A: 1}
	})

	var x1 *UseCaseA
	err := Resolve(&x1)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x1)

	var x2 *UseCaseA
	err = Resolve(&x2)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, x2)

	expect.Same(x1, x2)
}

func TestGlobal_Value(t *testing.T) {
	expect := assert.New(t)

	x1 := &UseCaseA{A: 1}
	Value(x1)

	var d1 *UseCaseA
	err := Resolve(&d1)
	expect.NoError(err)
	expect.IsType(&UseCaseA{}, d1)
	expect.Same(x1, d1)
}

func TestGlobal_Inject(t *testing.T) {
	expect := assert.New(t)

	Register(func() Database {
		return &DatabaseImpl{A: 1}
	}).NamedRegister("d1", func() Database {
		return &DatabaseImpl{A: 2}
	})

	repo := &Repo{}
	err := Inject(repo)

	expect.NoError(err)
	expect.NotNil(repo.Db1)
	expect.NotNil(repo.Db2)
	expect.Equal(1, repo.Db1.Read())
	expect.Equal(2, repo.Db2.Read())
}
