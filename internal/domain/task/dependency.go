// © 2026 Gabriel Pedreros — Todos los derechos reservados (ver LICENSE).
package task

// Dependency declara que una Task depende de otra.
type Dependency struct {
	TaskID      string
	DependsOnID string
}

// Graph agrupa tareas y dependencias con ordenación topológica.
type Graph struct {
	Tasks map[string]*Task
}

func NewGraph() *Graph {
	return &Graph{Tasks: make(map[string]*Task)}
}

func (g *Graph) Add(t Task) {
	g.Tasks[t.ID] = &t
}

// Ready devuelve las tareas cuyas dependencias están completadas.
func (g *Graph) Ready() []*Task {
	var ready []*Task
	for _, t := range g.Tasks {
		if t.Status != StatusPending && t.Status != StatusQueued {
			continue
		}
		ok := true
		for _, depID := range t.Dependencies {
			dep, exists := g.Tasks[depID]
			if !exists || dep.Status != StatusCompleted {
				ok = false
				break
			}
		}
		if ok {
			ready = append(ready, t)
		}
	}
	return ready
}

// TopologicalOrder devuelve las tareas en orden de dependencias (algoritmo de Kahn, mejor esfuerzo).
func (g *Graph) TopologicalOrder() []*Task {
	// Copia el grado de entrada por tarea.
	inDegree := make(map[string]int)
	for id := range g.Tasks {
		inDegree[id] = 0
	}
	for _, t := range g.Tasks {
		for _, dep := range t.Dependencies {
			if _, ok := g.Tasks[dep]; ok {
				inDegree[t.ID]++
			}
		}
	}
	var queue []*Task
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, g.Tasks[id])
		}
	}
	var order []*Task
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		for _, t := range g.Tasks {
			for _, dep := range t.Dependencies {
				if dep == n.ID {
					inDegree[t.ID]--
					if inDegree[t.ID] == 0 {
						queue = append(queue, t)
					}
				}
			}
		}
	}
	return order
}
