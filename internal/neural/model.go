// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package neural

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"
)

var (
	ErrNoLayers         = errors.New("neural: model has no layers")
	ErrInvalidModelFile = errors.New("neural: invalid model file")
)

type Layer interface {
	Forward(input *Tensor) (*Tensor, error)
	Backward(gradOutput *Tensor) (*Tensor, error)
	GetWeights() *Tensor
	GetBias() *Tensor
	GetGradWeights() *Tensor
	GetGradBias() *Tensor
	InputDim() int
	OutputDim() int
}

type Model struct {
	layers      []Layer
	activations []Activation
	inputDim    int
	outputDim   int
}

func NewModel(inputDim int) *Model {
	return &Model{inputDim: inputDim}
}

func (m *Model) AddDense(outputDim int, rng *rand.Rand) error {
	inputDim := m.inputDim
	if len(m.layers) > 0 {
		inputDim = m.layers[len(m.layers)-1].OutputDim()
	}
	dense, err := NewDense(inputDim, outputDim, rng)
	if err != nil {
		return err
	}
	m.layers = append(m.layers, dense)
	m.outputDim = outputDim
	return nil
}

func (m *Model) AddActivation(a Activation) {
	m.activations = append(m.activations, a)
}

func (m *Model) Forward(input *Tensor) (*Tensor, error) {
	if len(m.layers) == 0 {
		return nil, ErrNoLayers
	}
	current := input
	actIdx := 0
	for _, layer := range m.layers {
		var err error
		current, err = layer.Forward(current)
		if err != nil {
			return nil, err
		}
		if actIdx < len(m.activations) {
			current, err = m.activations[actIdx].Forward(current)
			if err != nil {
				return nil, err
			}
			actIdx++
		}
	}
	return current, nil
}

func (m *Model) Backward(gradOutput *Tensor) error {
	if len(m.layers) == 0 {
		return ErrNoLayers
	}
	current := gradOutput
	for i := len(m.layers) - 1; i >= 0; i-- {
		if i < len(m.activations) {
			var err error
			current, err = m.activations[i].Backward(current)
			if err != nil {
				return err
			}
		}
		var err error
		current, err = m.layers[i].Backward(current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) GetParamsAndGrads() (params, grads []*Tensor) {
	for _, layer := range m.layers {
		if w := layer.GetWeights(); w != nil {
			params = append(params, w)
			grads = append(grads, layer.GetGradWeights())
		}
		if b := layer.GetBias(); b != nil {
			params = append(params, b)
			grads = append(grads, layer.GetGradBias())
		}
	}
	return
}

func (m *Model) InputDim() int   { return m.inputDim }
func (m *Model) OutputDim() int  { return m.outputDim }
func (m *Model) Layers() []Layer { return m.layers }

type WeightData struct {
	Rows int       `json:"rows"`
	Cols int       `json:"cols"`
	Data []float64 `json:"data"`
}

type ModelData struct {
	InputDim  int                   `json:"input_dim"`
	OutputDim int                   `json:"output_dim"`
	Weights   map[string]WeightData `json:"weights"`
}

func SaveModel(path string, m *Model) error {
	md := ModelData{
		InputDim:  m.inputDim,
		OutputDim: m.outputDim,
		Weights:   make(map[string]WeightData),
	}
	for i, layer := range m.layers {
		if w := layer.GetWeights(); w != nil {
			md.Weights[weightsKey(i)] = WeightData{
				Rows: w.Rows, Cols: w.Cols, Data: w.Data,
			}
		}
		if b := layer.GetBias(); b != nil {
			md.Weights[biasKey(i)] = WeightData{
				Rows: b.Rows, Cols: b.Cols, Data: b.Data,
			}
		}
	}
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadModel(path string) (*ModelData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var md ModelData
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, ErrInvalidModelFile
	}
	return &md, nil
}

func RestoreModel(md *ModelData, layers []Layer) error {
	for i, layer := range layers {
		if wd, ok := md.Weights[weightsKey(i)]; ok {
			w := layer.GetWeights()
			if w != nil && w.Rows == wd.Rows && w.Cols == wd.Cols {
				copy(w.Data, wd.Data)
			}
		}
		if bd, ok := md.Weights[biasKey(i)]; ok {
			b := layer.GetBias()
			if b != nil && b.Rows == bd.Rows && b.Cols == bd.Cols {
				copy(b.Data, bd.Data)
			}
		}
	}
	return nil
}

func weightsKey(i int) string { return "dense_" + strconv.Itoa(i) + "_weights" }
func biasKey(i int) string    { return "dense_" + strconv.Itoa(i) + "_bias" }
