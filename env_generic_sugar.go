//go:build go1.18
// +build go1.18

package fixenv

func Cache[TEnv Env, TRes any](env TEnv, params any, opt *FixtureOptions, f func() (TRes, error)) TRes {
	callbackResult := env.Cache(params, opt, func() (res interface{}, err error) {
		return f()
	})

	var res TRes
	if callbackResult != nil {
		res = callbackResult.(TRes)
	}
	return res
}
