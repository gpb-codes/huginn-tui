package cron

import "time"

// Dream — cada 7 días converge memorias dispersas (Mimo)
type Dream struct{ Interval time.Duration }

func NewDream() *Dream { return &Dream{Interval: 7 * 24 * time.Hour} }

// Distill — cada 30 días solidifica workflows repetidos en skills
type Distill struct{ Interval time.Duration }

func NewDistill() *Distill { return &Distill{Interval: 30 * 24 * time.Hour} }
