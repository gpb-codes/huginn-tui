// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package routing

import (
	"strings"

	neural "huginn/internal/neural"
)

type TaskFeatures struct {
	Complexity        float64
	CodeRelated       float64
	TerminalRequired  float64
	ResearchRequired  float64
	ReasoningRequired float64
	ContextSize       float64
	HistoricalSuccess float64
}

func (tf TaskFeatures) ToTensor() *neural.Tensor {
	data := []float64{
		tf.Complexity,
		tf.CodeRelated,
		tf.TerminalRequired,
		tf.ResearchRequired,
		tf.ReasoningRequired,
		tf.ContextSize,
		tf.HistoricalSuccess,
	}
	t, _ := neural.FromData(data, 1, 7)
	return t
}

type FeatureExtractor struct{}

func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

func (fe *FeatureExtractor) Extract(title, description string) TaskFeatures {
	combined := strings.ToLower(title + " " + description)
	return TaskFeatures{
		Complexity:        fe.estimateComplexity(combined),
		CodeRelated:       fe.estimateCodeRelated(combined),
		TerminalRequired:  fe.estimateTerminal(combined),
		ResearchRequired:  fe.estimateResearch(combined),
		ReasoningRequired: fe.estimateReasoning(combined),
		ContextSize:       fe.estimateContextSize(description),
		HistoricalSuccess: 0.5,
	}
}

func (fe *FeatureExtractor) estimateComplexity(text string) float64 {
	score := 0.3
	complexWords := []string{"sistema", "arquitectura", "complejo", "multi", "integración", "distribuido", "escalable", "optimizar", "refactor", "migrar"}
	for _, w := range complexWords {
		if strings.Contains(text, w) {
			score += 0.07
		}
	}
	if len(text) > 200 {
		score += 0.1
	}
	return clamp01(score)
}

func (fe *FeatureExtractor) estimateCodeRelated(text string) float64 {
	score := 0.2
	codeWords := []string{"código", "code", "implementar", "programar", "función", "clase", "método", "api", "endpoint", "handler", "bug", "error", "fix", "test", "compilar", "build", "deploy"}
	for _, w := range codeWords {
		if strings.Contains(text, w) {
			score += 0.08
		}
	}
	return clamp01(score)
}

func (fe *FeatureExtractor) estimateTerminal(text string) float64 {
	score := 0.1
	termWords := []string{"terminal", "consola", "shell", "bash", "powershell", "ejecutar", "instalar", "npm", "go", "pip", "docker", "git", "commit", "push", "deploy"}
	for _, w := range termWords {
		if strings.Contains(text, w) {
			score += 0.1
		}
	}
	return clamp01(score)
}

func (fe *FeatureExtractor) estimateResearch(text string) float64 {
	score := 0.15
	resWords := []string{"investigar", "research", "analizar", "documentación", "docs", "referencia", "comparar", "evaluar", "estudiar", "buscar", "fuentes"}
	for _, w := range resWords {
		if strings.Contains(text, w) {
			score += 0.1
		}
	}
	return clamp01(score)
}

func (fe *FeatureExtractor) estimateReasoning(text string) float64 {
	score := 0.2
	reasonWords := []string{"razonar", "analizar", "explicar", "por qué", "causa", "efecto", "estrategia", "diseñar", "planificar", "decidir", "argumentar", "lógica"}
	for _, w := range reasonWords {
		if strings.Contains(text, w) {
			score += 0.08
		}
	}
	return clamp01(score)
}

func (fe *FeatureExtractor) estimateContextSize(text string) float64 {
	n := float64(len(text))
	if n < 50 {
		return 0.2
	}
	if n < 200 {
		return 0.4
	}
	if n < 500 {
		return 0.6
	}
	if n < 1000 {
		return 0.8
	}
	return 1.0
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
