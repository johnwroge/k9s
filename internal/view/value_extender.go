// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"log/slog"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// ValueExtender adds values actions to a given viewer.
type ValueExtender struct {
	ResourceViewer
}

// NewValueExtender returns a new extender.
func NewValueExtender(r ResourceViewer) ResourceViewer {
	p := ValueExtender{ResourceViewer: r}
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(func(*App, ui.Tabular, *client.GVR, string) {
		p.valuesCmd(nil)
	})

	return &p
}

func (v *ValueExtender) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyV, ui.NewKeyAction("Values", v.valuesCmd, true))
}

func (v *ValueExtender) valuesCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := v.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	showValues(v.defaultCtx(), v.App(), path, v.GVR())
	return nil
}

func (v *ValueExtender) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, v.App().factory)
}

func showValues(ctx context.Context, app *App, path string, gvr *client.GVR) {
	vm := model.NewValues(gvr, path)
	if err := vm.Init(app.factory); err != nil {
		app.Flash().Errf("Initializing the values model failed: %s", err)
	}

	toggleValuesCmd := func(*tcell.EventKey) *tcell.EventKey {
		if err := vm.ToggleValues(); err != nil {
			app.Flash().Errf("Values toggle failed: %s", err)
			return nil
		}

		if err := vm.Refresh(ctx); err != nil {
			slog.Error("Values viewer refresh failed", slogs.Error, err)
			return nil
		}

		app.Flash().Infof("Values toggled")
		return nil
	}

	v := NewLiveView(app, "Values", vm)
	v.actions.Add(ui.KeyV, ui.NewKeyAction("Toggle All Values", toggleValuesCmd, true))
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}
}
