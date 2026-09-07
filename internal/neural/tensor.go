// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

var (
	ErrEmptyTensor       = errors.New("neural: empty tensor")
	ErrDimensionMismatch = errors.New("neural: dimension mismatch")
	ErrInvalidShape      = errors.New("neural: invalid shape")
	ErrNonPositiveDim    = errors.New("neural: dimension must be positive")
)

type Tensor struct {
	Data []float64
	Rows int
	Cols int
}

func New(rows, cols int) (*Tensor, error) {
	if rows <= 0 || cols <= 0 {
		return nil, ErrNonPositiveDim
	}
	return &Tensor{
		Data: make([]float64, rows*cols),
		Rows: rows,
		Cols: cols,
	}, nil
}

func Zeros(rows, cols int) (*Tensor, error) {
	return New(rows, cols)
}

func Ones(rows, cols int) (*Tensor, error) {
	t, err := New(rows, cols)
	if err != nil {
		return nil, err
	}
	for i := range t.Data {
		t.Data[i] = 1.0
	}
	return t, nil
}

func Random(rows, cols int, rng *rand.Rand) (*Tensor, error) {
	t, err := New(rows, cols)
	if err != nil {
		return nil, err
	}
	for i := range t.Data {
		t.Data[i] = rng.NormFloat64()
	}
	return t, nil
}

func RandomUniform(rows, cols int, low, high float64, rng *rand.Rand) (*Tensor, error) {
	t, err := New(rows, cols)
	if err != nil {
		return nil, err
	}
	for i := range t.Data {
		t.Data[i] = low + rng.Float64()*(high-low)
	}
	return t, nil
}

func FromData(data []float64, rows, cols int) (*Tensor, error) {
	if rows <= 0 || cols <= 0 {
		return nil, ErrNonPositiveDim
	}
	if len(data) != rows*cols {
		return nil, ErrDimensionMismatch
	}
	out := &Tensor{
		Data: make([]float64, len(data)),
		Rows: rows,
		Cols: cols,
	}
	copy(out.Data, data)
	return out, nil
}

func (t *Tensor) At(r, c int) float64 {
	return t.Data[r*t.Cols+c]
}

func (t *Tensor) Set(r, c int, v float64) {
	t.Data[r*t.Cols+c] = v
}

func (t *Tensor) Clone() *Tensor {
	out := &Tensor{
		Data: make([]float64, len(t.Data)),
		Rows: t.Rows,
		Cols: t.Cols,
	}
	copy(out.Data, t.Data)
	return out
}

func (t *Tensor) Shape() (int, int) {
	return t.Rows, t.Cols
}

func (t *Tensor) Size() int {
	return len(t.Data)
}

func (t *Tensor) IsEmpty() bool {
	return len(t.Data) == 0
}

func (t *Tensor) validate() error {
	if t.IsEmpty() {
		return ErrEmptyTensor
	}
	if t.Rows <= 0 || t.Cols <= 0 {
		return ErrInvalidShape
	}
	return nil
}

func (t *Tensor) Mul(other *Tensor) (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	if err := other.validate(); err != nil {
		return nil, err
	}
	if t.Cols != other.Rows {
		return nil, fmt.Errorf("%w: (%d,%d) × (%d,%d)", ErrDimensionMismatch, t.Rows, t.Cols, other.Rows, other.Cols)
	}
	out, _ := New(t.Rows, other.Cols)
	for i := 0; i < t.Rows; i++ {
		for k := 0; k < t.Cols; k++ {
			aik := t.At(i, k)
			if aik == 0 {
				continue
			}
			for j := 0; j < other.Cols; j++ {
				out.Data[i*other.Cols+j] += aik * other.At(k, j)
			}
		}
	}
	return out, nil
}

func (t *Tensor) Add(other *Tensor) (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	if err := other.validate(); err != nil {
		return nil, err
	}
	if t.Rows != other.Rows || t.Cols != other.Cols {
		return nil, fmt.Errorf("%w: (%d,%d) + (%d,%d)", ErrDimensionMismatch, t.Rows, t.Cols, other.Rows, other.Cols)
	}
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] += other.Data[i]
	}
	return out, nil
}

func (t *Tensor) Sub(other *Tensor) (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	if err := other.validate(); err != nil {
		return nil, err
	}
	if t.Rows != other.Rows || t.Cols != other.Cols {
		return nil, fmt.Errorf("%w: (%d,%d) - (%d,%d)", ErrDimensionMismatch, t.Rows, t.Cols, other.Rows, other.Cols)
	}
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] -= other.Data[i]
	}
	return out, nil
}

func (t *Tensor) Scale(s float64) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= s
	}
	return out
}

func (t *Tensor) Transpose() (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	out, _ := New(t.Cols, t.Rows)
	for i := 0; i < t.Rows; i++ {
		for j := 0; j < t.Cols; j++ {
			out.Data[j*t.Rows+i] = t.At(i, j)
		}
	}
	return out, nil
}

func (t *Tensor) Map(fn func(float64) float64) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] = fn(out.Data[i])
	}
	return out
}

func (t *Tensor) ArgMax() (int, int) {
	if t.IsEmpty() {
		return 0, 0
	}
	best := 0
	bestVal := t.Data[0]
	for i := 1; i < len(t.Data); i++ {
		if t.Data[i] > bestVal {
			bestVal = t.Data[i]
			best = i
		}
	}
	return best / t.Cols, best % t.Cols
}

func (t *Tensor) ArgMaxRow(row int) int {
	if t.IsEmpty() || row < 0 || row >= t.Rows {
		return 0
	}
	start := row * t.Cols
	best := 0
	bestVal := t.Data[start]
	for j := 1; j < t.Cols; j++ {
		if t.Data[start+j] > bestVal {
			bestVal = t.Data[start+j]
			best = j
		}
	}
	return best
}

func Softmax(t *Tensor) (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	out := t.Clone()
	for i := 0; i < t.Rows; i++ {
		max := t.Data[i*t.Cols]
		for j := 1; j < t.Cols; j++ {
			if t.Data[i*t.Cols+j] > max {
				max = t.Data[i*t.Cols+j]
			}
		}
		sum := 0.0
		for j := 0; j < t.Cols; j++ {
			out.Data[i*t.Cols+j] = math.Exp(t.Data[i*t.Cols+j] - max)
			sum += out.Data[i*t.Cols+j]
		}
		for j := 0; j < t.Cols; j++ {
			out.Data[i*t.Cols+j] /= sum
		}
	}
	return out, nil
}

func (t *Tensor) Sum() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v
	}
	return s
}

func (t *Tensor) Mean() float64 {
	if t.IsEmpty() {
		return 0
	}
	return t.Sum() / float64(len(t.Data))
}

func (t *Tensor) AddScalar(s float64) *Tensor {
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] += s
	}
	return out
}

func (t *Tensor) Hadamard(other *Tensor) (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	if err := other.validate(); err != nil {
		return nil, err
	}
	if t.Rows != other.Rows || t.Cols != other.Cols {
		return nil, fmt.Errorf("%w: hadamard (%d,%d) × (%d,%d)", ErrDimensionMismatch, t.Rows, t.Cols, other.Rows, other.Cols)
	}
	out := t.Clone()
	for i := range out.Data {
		out.Data[i] *= other.Data[i]
	}
	return out, nil
}

func (t *Tensor) Reshape(rows, cols int) (*Tensor, error) {
	if rows*cols != len(t.Data) {
		return nil, fmt.Errorf("%w: reshape (%d,%d) from %d elements", ErrDimensionMismatch, rows, cols, len(t.Data))
	}
	out := t.Clone()
	out.Rows = rows
	out.Cols = cols
	return out, nil
}

func (t *Tensor) L2Norm() float64 {
	s := 0.0
	for _, v := range t.Data {
		s += v * v
	}
	return math.Sqrt(s)
}

func (t *Tensor) SumRows() (*Tensor, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	out, _ := New(t.Rows, 1)
	for i := 0; i < t.Rows; i++ {
		sum := 0.0
		for j := 0; j < t.Cols; j++ {
			sum += t.At(i, j)
		}
		out.Data[i] = sum
	}
	return out, nil
}
