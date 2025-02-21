package schedule

// Schedueler allows for an execution of pipelines based
// of a delay
type Schedueler struct {
	cron   string
	Hour   int64
	Minute int64
	Second int64
	Day    int64
	Month  int64
	Year   int64
}

type SchedulerStarter struct {
	schedueler *Schedueler
}

type ScheduelerBuilder struct {
	schedueler *Schedueler
	count      int64
}

type ScheduelerBuilderTransitionner struct {
	schedueler *Schedueler
	count      int64
}

func (s *Schedueler) Scheduele() *SchedulerStarter {
	return &SchedulerStarter{schedueler: s}
}

func (s *SchedulerStarter) Every(num int64) *ScheduelerBuilder {
	return &ScheduelerBuilder{count: num, schedueler: s.schedueler}
}

func (sb *ScheduelerBuilder) Minute() *ScheduelerBuilderTransitionner {
	sb.schedueler.Minute = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sb *ScheduelerBuilder) Second() *ScheduelerBuilderTransitionner {
	sb.schedueler.Second = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sb *ScheduelerBuilder) Hour() *ScheduelerBuilderTransitionner {
	sb.schedueler.Hour = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sb *ScheduelerBuilder) Day() *ScheduelerBuilderTransitionner {
	sb.schedueler.Hour = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sb *ScheduelerBuilder) Month() *ScheduelerBuilderTransitionner {
	sb.schedueler.Month = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sb *ScheduelerBuilder) Year() *ScheduelerBuilderTransitionner {
	sb.schedueler.Year = sb.count
	return &ScheduelerBuilderTransitionner{schedueler: sb.schedueler}
}

func (sbt *ScheduelerBuilderTransitionner) And(num int64) *ScheduelerBuilder {
	return &ScheduelerBuilder{schedueler: sbt.schedueler, count: num}
}

func NewSchedueler() *Schedueler {
	return &Schedueler{}
}

func test() {
    s := NewSchedueler()
    s.Scheduele().Every(5).Month().And(12).Day()
}
