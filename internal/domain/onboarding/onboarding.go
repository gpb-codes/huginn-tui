package onboarding

import "strings"

// Secciones del onboarding — orden fijo: perfil tecnico > desarrollo > git > ia > preferencias > resumen
type Step string

const (
	StepVaultChoice   Step = "vault_choice"
	StepVaultConfirm  Step = "vault_confirm"
	StepMode          Step = "mode" // manual / automatico / omitir
	StepTechnical     Step = "technical"
	StepDevelopment   Step = "development"
	StepGit           Step = "git"
	StepAI            Step = "ai"
	StepPreferences   Step = "preferences"
	StepSummary       Step = "summary"
)

var AllSteps = []Step{
	StepVaultChoice, StepVaultConfirm, StepMode,
	StepTechnical, StepDevelopment, StepGit, StepAI, StepPreferences, StepSummary,
}

// Datos recopilados — cada campo admite Skip (vacío)
type TechnicalProfile struct {
	Languages      []string `json:"languages"`
	PrimaryLang    string   `json:"primary_lang"`
	Frameworks     []string `json:"frameworks"`
	Stack          string   `json:"stack"`
	Editor         string   `json:"editor"`
	Tools          []string `json:"tools"`
	OS             string   `json:"os"`
	Skipped        bool     `json:"skipped"`
}

type DevelopmentPrefs struct {
	ProjectStructure string `json:"project_structure"` // ej: modular, monorepo
	Methodology      string `json:"methodology"`       // agile, trunk-based
	Architecture     string `json:"architecture"`       // hexagonal, clean
	FileOrganization string `json:"file_organization"`
	TestingStrategy  string `json:"testing_strategy"`
	TestingLevel     string `json:"testing_level"`
	NamingConvention string `json:"naming_convention"`
	Documentation    string `json:"documentation"`
	Skipped          bool   `json:"skipped"`
}

type GitPrefs struct {
	Workflow      string `json:"workflow"`       // feature branches, git flow, trunk
	BranchStrategy string `json:"branch_strategy"`
	CommitConvention string `json:"commit_convention"` // conventional commits
	PRStrategy    string `json:"pr_strategy"`
	AgentRepoRules string `json:"agent_repo_rules"`
	Skipped       bool   `json:"skipped"`
}

type AIPrefs struct {
	Providers      []string `json:"providers"`
	LocalOrCloud   string   `json:"local_or_cloud"`
	AutonomyLevel  string   `json:"autonomy_level"` // conservador / autonomo
	NeedsApproval  bool     `json:"needs_approval"`
	DelegatedTasks []string `json:"delegated_tasks"`
	ForbiddenActions []string `json:"forbidden_actions"`
	Skipped        bool     `json:"skipped"`
}

type GeneralPrefs struct {
	AnswerStyle      string `json:"answer_style"` // conciso / detallado
	ExplainDecisions bool   `json:"explain_decisions"`
	ProposeAlternatives bool `json:"propose_alternatives"`
	AgentStyle       string `json:"agent_style"` // conservador / autonomo
	GlobalRules      []string `json:"global_rules"`
	Skipped          bool     `json:"skipped"`
}

type Result struct {
	VaultPath   string           `json:"vault_path"`
	Mode        string           `json:"mode"` // manual / automatico / omitir
	Technical   TechnicalProfile `json:"technical"`
	Development DevelopmentPrefs `json:"development"`
	Git         GitPrefs         `json:"git"`
	AI          AIPrefs          `json:"ai"`
	General     GeneralPrefs     `json:"general"`
}

// IsComplete verifica si al menos Vault + una seccion tiene datos
func (r Result) IsComplete() bool {
	return strings.TrimSpace(r.VaultPath) != ""
}

// NeedsSecrets detecta si hay intento de guardar secretos en markdown
func NeedsSecrets(s string) bool {
	low := strings.ToLower(s)
	secrets := []string{"api_key", "api-key", "apikey", "password", "secret", "token", "sk-", "bearer"}
	for _, k := range secrets {
		if strings.Contains(low, k) {
			return true
		}
	}
	return false
}
