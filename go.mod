module github.com/laupse/dagger-kapp

go 1.19

replace github.com/docker/docker => github.com/docker/docker v20.10.3-0.20220414164044-61404de7df1a+incompatible

require (
	github.com/Khan/genqlient v0.5.0
	go.dagger.io/dagger v0.2.35-0.20220930232833-f63275b72bb6
)

require github.com/vektah/gqlparser/v2 v2.5.1 // indirect
