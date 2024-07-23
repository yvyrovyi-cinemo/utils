package circlebuff

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCircleBuff_Avg(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 60*time.Second)

	calculator.AddValue(1.0)
	res, err := calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 1.0, res)

	calculator.AddValue(2)
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 1.5, res)

	calculator.AddValue(3)
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 2.0, res)

	calculator.AddValue(4)
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 3.0, res)

	calculator.AddValue(5)
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 4.0, res)
}

func TestCircleBuff_Avg_Expiry(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 30*time.Second)

	now := time.Now()

	calculator.AddValueWithTm(1.0, now.Add(-5*time.Second))
	res, err := calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 1.0, res)

	calculator.AddValueWithTm(2, now.Add(-5*time.Second))
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 1.5, res)

	calculator.AddValueWithTm(3, now.Add(-5*time.Second))
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 2.0, res)

	calculator.AddValueWithTm(4, now.Add(-5*time.Second))
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 3.0, res)

	calculator.AddValueWithTm(5, now.Add(-50*time.Second))
	res, err = calculator.Avg()
	require.NoError(t, err)
	require.Equal(t, 3.5, res)
}

func TestCircleBuff_Avg_NoValues(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 30*time.Second)

	now := time.Now()

	calculator.AddValueWithTm(1.0, now.Add(-5*time.Minute))
	calculator.AddValueWithTm(1.0, now.Add(-3*time.Minute))

	res, err := calculator.Avg()
	require.True(t, errors.Is(err, ErrNoValues))
	require.Equal(t, 0.0, res)
}

func TestCircleBuff_Sum(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 30*time.Second)

	now := time.Now()

	calculator.AddValueWithTm(1, now.Add(-5*time.Second))
	res, err := calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 1, res)

	calculator.AddValueWithTm(2, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 3, res)

	calculator.AddValueWithTm(3, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 6, res)

	calculator.AddValueWithTm(4, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 9, res)

	calculator.AddValueWithTm(5, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 12, res)
}

func TestCircleBuff_Sum_Expiry(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 30*time.Second)

	now := time.Now()

	calculator.AddValueWithTm(1, now.Add(-5*time.Second))
	res, err := calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 1, res)

	calculator.AddValueWithTm(2, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 3, res)

	calculator.AddValueWithTm(3, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 6, res)

	calculator.AddValueWithTm(4, now.Add(-5*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 9, res)

	calculator.AddValueWithTm(5, now.Add(-50*time.Second))
	res, err = calculator.Sum()
	require.NoError(t, err)
	require.Equal(t, 7, res)
}

func TestCircleBuff_Sum_NoValues(t *testing.T) {
	t.Parallel()

	calculator := New[int](3, 30*time.Second)

	now := time.Now()

	calculator.AddValueWithTm(1, now.Add(-50*time.Second))
	calculator.AddValueWithTm(1, now.Add(-50*time.Second))

	res, err := calculator.Sum()
	require.True(t, errors.Is(err, ErrNoValues))
	require.Equal(t, 0, res)
}
