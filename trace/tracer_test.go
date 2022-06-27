package trace

import (
	"fmt"
	"io"
)

type Tracer interface {
	Trace(...interface{}) //...は可変長引数   インターフェースと同じ名前のメソッド
}

type tracer struct {
	out io.Writer
}

type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {} //nilTracer構造体には、何も処理を行わないTraceメソッドが定義されている

//OffはTraceメソッドの呼び出しを無視するTracerを返します。
func Off() Tracer {
	return &nilTracer{}
}

func (t *tracer) Trace(a ...interface{}) { //tracer構造体に対して実装しているメソッド
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}

func New(w io.Writer) Tracer { //Trace interfaceに合致したオブジェクトとして受け取る  プライベートなtracer型については関知しない
	return &tracer{out: w}
}

//Accept Interfaces, Return Structs  ... 返り値としては具体的な型を返すけど，格納するインスタンスや関数の引数などは interface 型で受け入れる
