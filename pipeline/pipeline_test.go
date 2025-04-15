package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/Cyber-cicco/jerminal/config"
	"github.com/google/uuid"
)

func _test_getPipeline(agentId string) *Pipeline {
	return &Pipeline{
		Agent: &config.Agent{
			Identifier: agentId,
		},
		Name:          "test",
		mainDirectory: "./test",
		directory:     "./test",
		Id:            uuid.New(),
		TimeRan:       0,
		events:        []pipelineEvents{},
		Inerror:       false,
        PipelineParams: PipelineParams{
            params: make(map[Key]interface{}),
        },
		Diagnostic:    &Diagnostic{},
		globalState: config.GetStateCustomConf(&config.Config{
			AgentDir:             "./test/agent",
			PipelineDir:          "./test/pipeline",
			JerminalResourcePath: "../resources/jerminal.json",
			AgentResourcePath:    "../resources/agents.json",
		}),
	}
}

func TestPipelineExecution1(t *testing.T) {
	actual := 0
	expected := 46

	p := setPipelineWithState("test",
		Agent("test"),
		config.GetStateCustomConf(&config.Config{
			AgentDir:             "./test/agent",
			PipelineDir:          "./test/pipeline",
			JerminalResourcePath: "../resources/jerminal.json",
			AgentResourcePath:    "../resources/agents.json",
		}),
		Stages("stages1",
			Stage("s1s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
			Stage("s1s2",
				ExecDefer(
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					}),
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual *= actual // 16
						return nil
					}),
				),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
		),
		Stages("stages2",
			Stage("s2s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					}),
				),
			),
			Stage("s2s2",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual++
						return errors.New("test")
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return errors.New("test")
					}),
				),
			).DontStopIfErr(),
			Stage("s2s3",
				Defer(func(p *Pipeline, ctx context.Context) error {
					actual *= 2
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
		),
		Post(
			Success(func(p *Pipeline, ctx context.Context) error {
				t.Log("in success")
				actual++
				return nil
			}),
			Failure(func(p *Pipeline, ctx context.Context) error {
				t.Log("in failure")
				actual--
				return nil
			}),
			Always(func(p *Pipeline, ctx context.Context) error {
				t.Log("in always")
				actual++
				return nil
			}),
		),
	)

	err := p.ExecutePipeline(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if p.Inerror {
		t.Fatalf("Pipeline should not have been in error")
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestPipelineExecution2(t *testing.T) {
	actual := 0
	expected := 21

	p := setPipelineWithState("test",
		Agent("test"),
		config.GetStateCustomConf(&config.Config{
			AgentDir:             "./test/agent",
			PipelineDir:          "./test/pipeline",
			JerminalResourcePath: "../resources/jerminal.json",
			AgentResourcePath:    "../resources/agents.json",
		}),
		Stages("stages1",
			Stage("s1s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
			Stage("s1s2",
				ExecDefer(
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					}),
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual *= actual // 16
						return nil
					}),
				),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
		),
		Stages("stages2",
			Stage("s2s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return nil
					}),
				),
			),
			Stage("s2s2",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual++
						return errors.New("test")
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual++
						return errors.New("test")
					}),
				),
			),
			Stage("s2s3",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual++
					return nil
				}),
			),
		),
		Post(
			Success(func(p *Pipeline, ctx context.Context) error {
				t.Log("in success")
				actual++
				return nil
			}),
			Failure(func(p *Pipeline, ctx context.Context) error {
				t.Log("in failure")
				actual--
				return nil
			}),
			Always(func(p *Pipeline, ctx context.Context) error {
				t.Log("in always")
				actual++
				return nil
			}),
		),
	)

	err := p.ExecutePipeline(context.Background())

	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

	if !p.Inerror {
		t.Fatalf("Pipeline should have been in error")
	}

	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}

func TestPipelineParams(t *testing.T) {
	actualKey := Key("actual")
	p := setPipelineWithState("test",
		Agent("test"),
		config.GetStateCustomConf(&config.Config{
			AgentDir:             "./test/agent",
			PipelineDir:          "./test/pipeline",
			JerminalResourcePath: "../resources/jerminal.json",
			AgentResourcePath:    "../resources/agents.json",
		}),
		Stages("stages1",
			Stage("s1s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					p.Put(actualKey, 1)
					return nil
				}),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual := p.MustGet(actualKey).(int)
					actual++
					p.Put(actualKey, actual)
					return nil
				}),
			),
			Stage("s1s2",
				ExecDefer(
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual++
						p.Put(actualKey, actual)
						return nil
					}),
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual *= actual
						p.Put(actualKey, actual)
						return nil
					}),
				),
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual := p.MustGet(actualKey).(int)
					actual++
					p.Put(actualKey, actual)
					return nil
				}),
			),
		),
		Stages("stages2",
			Stage("s2s1",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual := p.MustGet(actualKey).(int)
					actual++
					p.Put(actualKey, actual)
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual++
						p.Put(actualKey, actual)
						return nil
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual++
						p.Put(actualKey, actual)
						return nil
					}),
				),
			),
			Stage("s2s2",
				Exec(func(p *Pipeline, ctx context.Context) error {
					actual := p.MustGet(actualKey).(int)
					actual++
					p.Put(actualKey, actual)
					return nil
				}),
				ExecTryCatch(
					func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual++
						p.Put(actualKey, actual)
						return errors.New("test")
					},
					Exec(func(p *Pipeline, ctx context.Context) error {
						actual := p.MustGet(actualKey).(int)
						actual++
						p.Put(actualKey, actual)
						return errors.New("test")
					}),
				),
			),
			Stage("s2s3",
				Exec(func(p *Pipeline, ctx context.Context) error {
                    actual := p.MustGet(actualKey).(int)
                    actual++
                    p.Put(actualKey, actual)
					return nil
				}),
			),
		),
		Post(
			Success(func(p *Pipeline, ctx context.Context) error {
				t.Log("in success")
                actual := p.MustGet(actualKey).(int)
                actual++
                p.Put(actualKey, actual)
				return nil
			}),
			Failure(func(p *Pipeline, ctx context.Context) error {
				t.Log("in failure")
                actual := p.MustGet(actualKey).(int)
                actual--
                p.Put(actualKey, actual)
				return nil
			}),
			Always(func(p *Pipeline, ctx context.Context) error {
				t.Log("in always")
                actual := p.MustGet(actualKey).(int)
                actual++
                p.Put(actualKey, actual)
				return nil
			}),
		),
	)
	err := p.ExecutePipeline(context.Background())
    actual := p.MustGet(actualKey).(int)

	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

	if !p.Inerror {
		t.Fatalf("Pipeline should have been in error")
	}

	expected := 21
	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}

}
