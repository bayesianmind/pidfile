package pidfile

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestPidCreated(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "pidfile-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	pidpathbase := filepath.Join(dir, "pid")
	t.Run("noPidFile", func(t *testing.T) {
		pidpath := pidpathbase + ".1"
		isRunning, err := IsRunning(pidpath)
		require.NoError(t, err)
		assert.False(t, isRunning)
		Remove(pidpath)
	})

	t.Run("running", func(t *testing.T) {
		pidpath := pidpathbase + ".2"
		require.NoError(t, Write(pidpath))

		isRunning, err := IsRunning(pidpath)
		require.NoError(t, err)
		assert.True(t, isRunning)
	})

	t.Run("stalePid", func(t *testing.T) {
		pidpath := pidpathbase + ".3"
		err = WriteControl(pidpath, 999999, false)
		require.NoError(t, err)
		pid, err := pidfileContents(pidpath)
		require.NoError(t, err)
		err = Write(pidpath)
		require.Equal(t, ErrFileStale, err)
		pidafter, err := pidfileContents(pidpath)
		require.NoError(t, err)
		assert.Equal(t, pid, pidafter)
	})

	t.Run("runningOtherPid", func(t *testing.T) {
		pidpath := pidpathbase + ".4"
		require.NoError(t, Write(pidpath))
		pid, err := pidfileContents(pidpath)
		require.NoError(t, err)
		err = WriteControl(pidpath, 999999, true)
		require.Equal(t, ErrProcessRunning, err)
		pidafter, err := pidfileContents(pidpath)
		require.NoError(t, err)
		assert.Equal(t, pid, pidafter)
	})

	t.Run("parentDirDoesntExist", func(t *testing.T) {
		pidpath := filepath.Join(pidpathbase, "nonexisting", "pid")
		require.Error(t, Write(pidpath))
	})




}