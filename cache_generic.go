//go:build go1.18

package fixenv

// 	Cache(params interface{}, opt *FixtureOptions, f FixtureCallbackFunc) interface{}
// type FixtureCallbackFunc func() (res interface{}, err error)

// GetOrSet is generic wrap for env.GetOrSet
// is experimental and can change any time
func GetOrSet[Res any](env Env, params any, opt *FixtureOptions, f func() (res Res, err error)) Res {
	resI := env.Cache(params, opt, func() (res interface{}, err error) {
		return f()
	})

	var res Res
	if resI != nil {
		res = resI.(Res)
	}
	return res
}
