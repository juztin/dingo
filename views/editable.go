// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package views

import (
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"minty.io/dingo"
)

var (
	editTempl, _  = NewTmpl("_dingoedit_").Parse(editTemplate)
	editableViews = make(map[string]View)
	CanEdit       = func(ctx dingo.Context) bool { return true }
	EmptyTmpl     = "<!doctype html><head><title>Template Doesn't Exist</title></head>" +
		"<body>This template doesn't exist, or hasn't been created yet.</body></html>"
	UseCodeMirror = true
	CodeMirrorJS  = "/js/libs/codemirror.js"
	CodeMirrorCSS = "/css/codemirror.css"
)

// EditTemplateData is data passed to the edit template.
type EditTemplateData struct {
	URL, DoneURL, DingoVer       string
	HasViews, IsAction, WasSaved bool
	Error                        error
	Views                        map[string]View
	Content                      []byte
	Stylesheets, Scripts         string
}

func editViewData(ctx dingo.Context, v View) EditTemplateData {
	d := new(EditTemplateData)
	d.DingoVer = dingo.VERSION
	d.DoneURL = ctx.URL.Path
	d.Content = []byte(v.Data(ctx))
	d.Stylesheets = codeMirrorCSS()
	d.Scripts = codeMirrorJS()

	return *d
}
func editCtxData(ctx dingo.Context) EditTemplateData {
	d := new(EditTemplateData)
	d.DingoVer = dingo.VERSION
	d.URL = ctx.URL.Path
	d.Views = editableViews
	d.HasViews = true
	d.Content = []byte("")
	d.Stylesheets = codeMirrorCSS()
	d.Scripts = codeMirrorJS()

	return *d
}

// Editable view wraps a view to be edited.
type EditableView struct {
	View
	tmpl *template.Template
}

// Editable returns a wrapped view that can be edited.
func Editable(view View) View {
	e := new(EditableView)
	e.View = view
	e.tmpl = editTempl
	// add the view to the editable cache
	editableViews[view.Name()] = view
	Add(view.Name(), e)

	return e
}

// Execute invokes the editor for the wrapped view.
func (e *EditableView) Execute(ctx dingo.Context, data interface{}) error {
	ctx.ParseForm()
	if _, ok := ctx.Form["edit"]; !ok || !CanEdit(ctx) {
		return e.View.Execute(ctx, data)
	}

	if ctx.Method == "GET" {
		return e.tmpl.Execute(ctx.Response, editViewData(ctx, e.View))
	} else if ctx.Method != "POST" {
		http.Error(ctx.Response, "Invalid Method", http.StatusMethodNotAllowed)
		return nil
	}

	c := []byte(ctx.FormValue("content"))
	d := new(EditTemplateData)
	d.DingoVer = dingo.VERSION
	d.IsAction = true
	d.DoneURL = ctx.URL.Path
	d.Stylesheets = codeMirrorCSS()
	d.Scripts = codeMirrorJS()
	if err := e.View.Save(ctx, c); err != nil {
		d.Error = err
		d.Content = c
	} else {
		d.WasSaved = true
		d.Content = []byte(e.View.Data(ctx))
	}

	return e.tmpl.Execute(ctx.Response, d)
	// TODO check if this has been updated
	// https://groups.google.com/forum/?fromgroups#!topic/golang-nuts/7Ks1iq2s7FA
	// len(edit) == 0 vs edit == ""
}

// EditHandler is a dingo.Handler that edits/saves a given template.
func EditHandler(ctx dingo.Context) {
	ctx.ParseForm()
	if !CanEdit(ctx) {
		ctx.HttpError(401)
		return
	}

	var v View
	d := editCtxData(ctx)
	if n, ok := ctx.Form["name"]; !ok {
		editTempl.Execute(ctx.Response, d)
		return
	} else if v, ok = editableViews[n[0]]; !ok {
		d.Error = errors.New(fmt.Sprintf("Template name: `%s` does not exist.", n[0]))
		editTempl.Execute(ctx.Response, d)
		return
	}

	if ctx.Method == "POST" {
		d.IsAction = true
		c := []byte(ctx.FormValue("content"))
		if err := v.Save(ctx, c); err != nil {
			d.Error = err
			d.Content = c
		} else {
			d.WasSaved = true
			d.Content = []byte(v.Data(ctx))
		}
	} else {
		d.Content = []byte(v.Data(ctx))
	}

	editTempl.Execute(ctx.Response, d)
}

// AddEditableView adds a view to be edited.
func AddEditableView(name string) {
	if v := Get(name); v != nil {
		editableViews[v.Name()] = v
	}
}

func stylesheet(url string) string {
	return fmt.Sprintf("<link rel='stylesheet' href='%s'>\n", url)
}
func script(url string) string {
	return fmt.Sprintf("<script src='%s'></script>\n", url)
}
func codeMirrorCSS() string {
	if UseCodeMirror {
		return stylesheet(CodeMirrorCSS)
	}
	return ""
}
func codeMirrorJS() string {
	if UseCodeMirror {
		return script(CodeMirrorJS) +
			"<script>CodeMirror.fromTextArea(document.getElementById('code'), {mode: 'text/html', tabSize: 2, indentWithTabs: true, smartIndent: false})</script>"
	}
	return ""
}

var dingoSvg = "<?xml version='1.0' encoding='UTF-8' standalone='yes'?>\n" +
	"<svg\n" +
	"   xmlns:svg='http://www.w3.org/2000/svg'\n" +
	"   xmlns='http://www.w3.org/2000/svg'\n" +
	"   width='48px'\n" +
	"   height='48px'>\n" +
	"  <g>\n" +
	"    <path\n" +
	"       style='fill:rgb(225,225,245);stroke:#050550;stroke-width:2;stroke-linecap:round;stroke-linejoin:round;stroke-miterlimit:4;stroke-opacity:1;stroke-dasharray:none;display:inline;'\n" +
	"       d='m 42.294549,-62.200441 c 0,0 -6.16964,-28.204012 -12.33926,-34.37364 -6.16963,-6.169629 -13.22063,9.69513 -17.62751,27.322637 C 7.9208987,-51.623937 2.6326487,-2.7076093 2.6326487,4.3434007 m 18.7907403,-2.98946 c 0,0 -25.4348503,-5.582798 -50.49122,31.9754303 -8.80297,13.19521 -42.785291,22.0258 -52.75992,46.00626 -15.73762,37.835539 45.79665,43.504709 98.5045,34.403669 2.74411,-0.47382 8.89204,-2.23414 10.3297,0.79798 53.35553,67.07316 65.17939,-6.44735 70.19118,-18.183489 7.275291,-22.96091 -0.30712,-21.86765 -0.96662,-35.94326 -1.0425,-22.24987 -2.45753,-34.50509 -4.70845,-45.99889 -7.35718,-37.567848 -2.5841,-33.591279 -17.77685,-93.346197 -20.68008,-49.429534 -22.95496,-4.678901 -34.23453,26.481125 -14.98332,41.391674 -18.08779,53.8073717 -18.08779,53.8073717 z'\n" +
	"       transform='matrix(0.22466416,0,0,0.19324629,23.29133,20.062688)' />\n" +
	"    <g\n" +
	"       transform='matrix(0.22466416,0,0,0.19324629,-6.4132665,-3.7095701)'\n" +
	"       style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;letter-spacing:0px;word-spacing:0px;writing-mode:lr-tb;text-anchor:start;fill:#003380;fill-opacity:1;stroke:none;display:inline;font-family:Nanum Brush Script;stroke-opacity:1'>\n" +
	"      <path\n" +
	"         d='m 98.732515,165.63183 c -0.824106,5.48519 -1.082436,10.70474 -0.774991,15.65864 0.320235,4.86835 0.44979,9.83943 0.388664,14.91325 -0.07409,5.15965 -0.355957,10.24419 -0.845607,15.25366 -0.476859,4.92384 -0.816467,9.51769 -1.018825,13.78158 -0.334411,0.47563 -0.5531,1.05625 -0.656067,1.74186 -0.01734,0.69854 -0.272448,1.22987 -0.765311,1.59399 -1.465875,1.00675 -3.096682,1.94487 -4.892425,2.81437 -1.795813,0.86954 -3.750125,1.6276 -5.862941,2.27418 -0.737171,0.23983 -1.448561,0.30824 -2.134172,0.20521 -0.672789,-0.18869 -1.364865,-0.24885 -2.07623,-0.18047 -0.79712,0.0556 -1.643505,0.14749 -2.539158,0.27582 -0.809993,0.14125 -1.630626,0.0618 -2.461902,-0.23841 -0.904144,-0.39873 -1.960364,-0.95179 -3.168663,-1.65918 -1.195456,-0.79305 -2.37802,-1.67182 -3.547694,-2.63633 -1.169701,-0.96446 -2.240808,-2.00177 -3.213323,-3.11194 -0.886832,-1.09724 -1.503663,-2.24158 -1.850495,-3.43302 -0.20883,-5.02678 0.625692,-9.41482 2.50357,-13.16415 1.96356,-3.73636 4.511975,-6.99052 7.645252,-9.76249 0.407181,-0.37697 0.88077,-0.61256 1.420771,-0.70676 0.552836,-0.17981 1.069278,-0.40896 1.549328,-0.68745 0.480006,-0.27838 0.99001,-0.46468 1.530013,-0.55889 0.552834,-0.1798 1.165643,-0.17538 1.83843,0.0133 1.688318,0.42899 3.509443,1.14078 5.46338,2.13539 1.966752,0.90901 3.494348,1.53289 4.582793,1.87163 l 0.289709,-1.92836 c 0.334741,-2.22827 0.498107,-4.48234 0.490099,-6.76223 0.0905,-2.3526 0.181038,-4.70525 0.271612,-7.05798 0.0905,-2.35258 0.230328,-4.74166 0.419483,-7.16722 0.287662,-2.49823 0.796061,-5.00721 1.525197,-7.52693 0.188675,-0.67268 0.463092,-1.33257 0.823253,-1.97966 0.360083,-0.64693 0.677353,-1.30038 0.95181,-1.96035 0.23575,-0.40268 0.563682,-0.83543 0.983798,-1.29824 0.43291,-0.54835 0.876523,-0.87608 1.330841,-0.98319 0.454233,-0.10695 0.749976,-0.32543 0.887227,-0.65546 0.150042,-0.41556 0.390056,-0.55478 0.720043,-0.41766 0.329899,0.1373 0.439142,0.28517 0.327729,0.44361 -0.0258,0.1715 -0.07086,0.47147 -0.135198,0.8999 m -9.771662,59.7918 c 0.227306,-0.92985 0.426884,-2.25827 0.598733,-3.98527 0.270361,-1.79977 0.290081,-3.68105 0.05916,-5.64385 -0.230991,-1.96272 -0.817643,-3.89128 -1.759957,-5.78568 -0.856676,-1.88144 -2.249074,-3.40522 -4.177199,-4.57134 -0.955658,-0.0559 -2.095571,-0.0519 -3.419745,0.0121 -1.225644,-0.009 -2.477015,0.15375 -3.754115,0.48766 -1.277145,0.33402 -2.500752,0.89512 -3.670827,1.68331 -1.157235,0.70258 -2.138818,1.69442 -2.944752,2.97553 -0.746011,1.46547 -1.048597,3.47953 -0.907757,6.04219 0.153687,2.47702 1.177311,4.99707 3.070877,7.56014 1.16967,0.96451 2.470128,1.6419 3.901378,2.03217 1.444084,0.30463 2.907503,0.48067 4.390261,0.52812 1.568412,0.0604 3.076897,-0.0635 4.525459,-0.37178 1.534207,-0.29531 2.897034,-0.6164 4.088486,-0.96327'\n" +
	"         style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;writing-mode:lr-tb;text-anchor:start;font-family:Nanum Brush Script;fill-opacity:1;fill:#003380;stroke:none;stroke-opacity:1' />\n" +
	"      <path\n" +
	"         d='m 117.56736,194.968 c 0.6823,0.70423 0.9334,1.27159 0.75328,1.70207 -0.0935,0.43198 -0.3562,0.6012 -0.78809,0.50765 -1.12655,-0.0177 -2.38511,0.0925 -3.77568,0.33058 -1.30258,0.1529 -2.73513,0.30368 -4.29765,0.45236 -1.47453,0.0635 -3.03434,0.0389 -4.67942,-0.0737 -1.64511,-0.11253 -3.28679,-0.44176 -4.925058,-0.98769 -0.68233,-0.7041 -0.672775,-1.31069 0.02867,-1.81977 0.789448,-0.59425 1.409688,-1.45125 1.860718,-2.57102 0.27225,-0.77574 0.62706,-1.29021 1.06444,-1.54342 0.52539,-0.33837 1.04874,-0.54682 1.57005,-0.62535 0.9532,0.0151 2.11964,0.25015 3.49932,0.7052 1.46768,0.3699 2.67745,0.60565 3.62932,0.70725 1.4704,0.19659 2.67949,0.47566 3.62726,0.83724 0.9491,0.27504 1.76005,1.06792 2.43284,2.37861 m -3.00812,17.63481 c 0.48851,2.00132 0.84981,3.82723 1.08386,5.47775 0.32067,1.65196 0.51206,3.25851 0.57416,4.81964 0.15009,1.47591 0.21286,2.99376 0.18831,4.55353 0.0621,1.5612 0.0788,3.25167 0.0501,5.07142 -0.005,0.34665 -0.22827,0.73319 -0.66835,1.15961 -0.3521,0.34119 -0.61821,0.72705 -0.79833,1.15757 -0.35893,0.77447 -0.66973,1.2463 -0.93241,1.41549 -0.17606,0.17061 -0.5261,0.38179 -1.05011,0.63354 -1.22139,0.50085 -2.31619,1.22036 -3.28441,2.15854 -1.1006,-1.66418 -1.85047,-3.5829 -2.24961,-5.75614 -0.31251,-2.17183 -0.53699,-4.42898 -0.67344,-6.77145 -0.0498,-2.34104 -0.14294,-4.6828 -0.27938,-7.02527 -0.0485,-2.42769 -0.31489,-4.77218 -0.79932,-7.03346 -0.24087,-1.21723 -0.14125,-2.0391 0.29887,-2.4656 0.44009,-0.42641 0.92148,-0.7222 1.44414,-0.88737 0.60931,-0.1637 1.25719,-0.0235 1.94362,0.42067 0.77306,0.44561 1.50281,0.89049 2.18925,1.33464 0.60112,0.35623 1.16098,0.58174 1.67956,0.67654 0.60657,0.01 1.0344,0.36305 1.28346,1.06035'\n" +
	"         style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;writing-mode:lr-tb;text-anchor:start;font-family:Nanum Brush Script;fill-opacity:1;fill:#003380;stroke:none;stroke-opacity:1' />\n" +
	"      <path\n" +
	"         d='m 130.07654,219.71312 c 1.1903,-1.71009 2.36533,-3.50554 3.52508,-5.38634 1.22974,-1.9813 2.51744,-3.88498 3.86309,-5.71105 1.34563,-1.82597 2.7645,-3.48902 4.25662,-4.98916 1.4768,-1.58533 3.08479,-2.93008 4.82397,-4.03427 0.39594,-0.24699 0.77218,-0.35845 1.12872,-0.33436 0.4418,0.009 0.8607,-0.11017 1.25668,-0.35729 0.68243,-0.12225 1.27959,-0.22927 1.79146,-0.32107 0.51181,-0.0917 1.01158,-0.005 1.4993,0.25958 1.33188,0.55378 2.46901,1.4946 3.41141,2.82246 1.01236,1.22738 1.92737,2.64824 2.74502,4.2626 0.50937,0.87728 0.84052,1.74243 0.99345,2.59545 0.23816,0.83784 0.6699,1.77301 1.29523,2.80551 0.80229,1.52916 1.65492,3.09327 2.55789,4.69235 0.97291,1.49858 1.51168,3.0309 1.6163,4.59697 -0.009,0.44186 -0.0681,0.84868 -0.17748,1.22045 -0.12479,0.28654 -0.2769,0.66599 -0.45634,1.13835 -0.1095,0.37185 -0.33163,0.8519 -0.6664,1.44015 -0.33484,0.58832 -0.69258,1.04865 -1.07321,1.38098 -0.36542,0.41771 -0.61048,0.7698 -0.73519,1.05626 -0.0548,0.18594 -0.40043,0.46801 -1.03699,0.84621 -1.39952,-1.42205 -2.56601,-3.01797 -3.49949,-4.78774 -0.93354,-1.76971 -1.79702,-3.64004 -2.59046,-5.61102 -0.7935,-1.97089 -1.55196,-3.99212 -2.27539,-6.0637 -0.72347,-2.07148 -1.53988,-4.17037 -2.44921,-6.29668 -0.27766,-0.56652 -0.89331,-0.80837 -1.84696,-0.72556 -0.86838,0.0676 -1.83856,0.54968 -2.91053,1.44613 -1.002,0.79596 -2.02048,1.99104 -3.05545,3.58524 -1.03501,1.5943 -1.87323,3.5495 -2.51469,5.86561 -0.40727,1.65788 -0.75659,3.39339 -1.04795,5.20652 -0.30668,1.7279 -0.691,3.51371 -1.15298,5.35742 -0.39198,1.74319 -0.95457,3.51691 -1.68776,5.32119 -0.66319,1.70374 -1.56257,3.31771 -2.69815,4.84192 -0.3501,0.503 -0.64103,0.59917 -0.87279,0.28849 -0.23178,-0.31062 -0.71183,-0.53276 -1.44015,-0.66639 -0.54243,-0.0789 -0.87158,-0.19596 -0.98746,-0.35131 -0.0306,-0.17059 -0.25028,-0.6595 -0.65908,-1.46674 -0.83292,-1.69969 -1.75435,-3.64769 -2.7643,-5.844 -0.92464,-2.21153 -1.95753,-4.53577 -3.09865,-6.97272 -1.05582,-2.45216 -2.15429,-4.89672 -3.29542,-7.33367 -1.15641,-2.52217 -2.28224,-4.87378 -3.37751,-7.05482 l -1.18057,-2.1657 c 0.10947,-0.37175 0.35454,-0.72384 0.7352,-1.05626 0.46598,-0.3476 0.93961,-0.6526 1.42087,-0.91501 0.48127,-0.26229 1.04341,-0.31901 1.68643,-0.17017 0.64302,0.14896 1.13514,0.19283 1.47637,0.13162 0.93838,-0.16812 1.85582,0.28379 2.75233,1.35572 0.98181,1.05676 1.89681,2.47763 2.74501,4.2626 0.83291,1.69978 1.57608,3.6357 2.22952,5.80778 0.72345,2.07158 1.30365,4.08073 1.74061,6.02747'\n" +
	"         style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;writing-mode:lr-tb;text-anchor:start;font-family:Nanum Brush Script;fill-opacity:1;fill:#003380;stroke:none;stroke-opacity:1' />\n" +
	"      <path\n" +
	"         d='m 171.89107,181.23003 c 1.59257,1.54879 3.76696,3.92508 6.52318,7.12887 2.75621,3.20391 5.62107,6.7398 8.59459,10.60768 2.9735,3.86799 5.84791,7.83056 8.62323,11.88772 2.77528,4.05725 4.9391,7.73204 6.49147,11.0244 1.21967,2.5869 2.165,5.30314 2.83599,8.14872 0.71237,2.73027 0.44712,5.82565 -0.79574,9.28616 -0.69986,1.76723 -1.56318,3.89898 -2.58996,6.39526 -0.91148,2.53772 -2.05363,4.99258 -3.42646,7.36458 -0.64503,1.07064 -1.3477,2.12058 -2.10803,3.1498 -0.68198,0.99225 -1.41938,1.86692 -2.2122,2.62402 -0.83549,-0.75589 -1.24876,-1.32757 -1.23981,-1.71505 0.12428,-0.34605 0.1192,-0.96647 -0.0152,-1.86125 -0.0975,-0.8164 -0.40272,-1.8702 -0.91568,-3.1614 -0.43461,-1.32816 -0.4677,-2.41445 -0.0993,-3.25887 0.45124,-1.07511 1.08446,-2.47557 1.89967,-4.20137 0.81517,-1.72578 1.57268,-3.47228 2.27253,-5.23952 0.66286,-1.8456 1.18743,-3.57809 1.57373,-5.19749 0.38625,-1.61936 0.43601,-2.93635 0.14929,-3.95097 -0.50853,-1.48492 -1.15087,-3.05049 -1.92702,-4.69671 -0.81315,-1.72456 -1.74387,-3.3937 -2.79215,-5.00744 -1.00689,-1.72902 -2.07368,-3.38191 -3.20035,-4.95869 -1.16367,-1.6551 -2.33179,-3.11649 -3.50437,-4.38418 l -0.56782,-0.59463 c -0.12756,2.16815 -0.48134,4.05966 -1.06135,5.67453 -0.53859,1.4996 -1.32636,2.8771 -2.3633,4.13252 -0.47929,0.60928 -0.99553,1.14013 -1.54872,1.59257 -0.59017,0.3741 -1.16185,0.78737 -1.71504,1.23981 -0.55321,0.45249 -1.06945,0.98334 -1.54872,1.59256 -0.51625,0.53088 -1.00954,0.90719 -1.47988,1.12893 -1.05603,0.40212 -2.27839,0.45146 -3.66707,0.14802 -1.31028,-0.34035 -2.28406,-0.98311 -2.92133,-1.92829 -0.33711,-0.51174 -0.75261,-0.98655 -1.24651,-1.42443 -0.53085,-0.51621 -0.98331,-1.06941 -1.3574,-1.65961 -1.22862,-2.19936 -2.45055,-4.68936 -3.66579,-7.47002 -1.21522,-2.78058 -2.27367,-5.63511 -3.17535,-8.56362 -0.86022,-3.04376 -1.45056,-6.02319 -1.77102,-8.9383 -0.35739,-2.99338 -0.3133,-5.74495 0.13226,-8.2547 0.25306,-0.88578 0.56379,-1.75091 0.93217,-2.5954 0.33145,-0.92274 0.96751,-1.6059 1.90818,-2.0495 0.62714,-0.29561 1.06498,-0.78951 1.31355,-1.48168 0.32698,-0.72899 0.82251,-1.20217 1.48658,-1.41954 1.40656,-0.47147 2.78793,-0.69161 4.14411,-0.66039 1.31923,-0.047 2.66424,0.4686 4.03504,1.5469 m 0.38192,11.17424 c -0.67422,-1.02349 -1.40612,-2.06774 -2.19571,-3.13278 -0.78957,-1.0649 -1.39881,-1.54418 -1.82773,-1.43783 -0.35052,0.0695 -0.6456,0.25655 -0.88524,0.56111 -0.19821,0.18933 -0.35722,0.36012 -0.47705,0.51237 -0.83426,0.8725 -1.20205,2.53108 -1.10338,4.97573 0.14012,2.32941 0.48798,4.89618 1.04359,7.7003 0.59705,2.68887 1.2932,5.28305 2.08845,7.78255 0.7583,2.4212 1.4376,4.06516 2.0379,4.93188 0.71119,1.10197 1.56068,2.09079 2.54847,2.96645 0.95083,0.79735 1.99122,1.07334 3.12117,0.82799 0.35052,-0.0694 0.8231,-0.38805 1.41773,-0.9559 0.55767,-0.64617 1.11757,-1.38924 1.67971,-2.22923 0.52518,-0.9183 1.05036,-1.83663 1.57554,-2.75501 0.48821,-0.99668 0.85661,-1.8411 1.10519,-2.53325 0.12429,-0.34601 -0.12997,-1.08848 -0.76276,-2.22741 -0.59137,-1.2542 -1.41121,-2.68817 -2.45951,-4.30193 -0.96991,-1.65061 -2.09214,-3.42109 -3.36667,-5.31145 -1.19614,-1.92721 -2.37604,-3.7184 -3.5397,-5.37359'\n" +
	"         style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;writing-mode:lr-tb;text-anchor:start;font-family:Nanum Brush Script;fill-opacity:1;fill:#003380;stroke:none;stroke-opacity:1' />\n" +
	"      <path\n" +
	"         d='m 192.33627,173.34483 -0.3628,3.10429 c 0.49012,-0.49796 1.1984,-1.27898 2.12483,-2.34306 0.94069,-1.1857 1.95489,-2.20861 3.04258,-3.06874 0.92331,-0.50903 1.81815,-0.77465 2.68452,-0.79685 0.86639,-0.0221 1.82448,0.49094 2.87426,1.53906 1.56205,1.41661 3.28692,2.75964 5.17464,4.0291 1.90195,1.14783 3.63394,2.43 5.19597,3.84649 1.63002,1.36285 2.86638,2.92644 3.70907,4.69076 0.78893,1.69645 0.89251,3.71379 0.31073,6.05204 0.31782,3.61577 0.17004,6.99197 -0.44334,10.12863 -0.59917,3.01499 -1.81097,5.46475 -3.63541,7.34929 -1.19838,1.27906 -2.82837,2.29171 -4.88998,3.03794 -2.11535,0.67829 -4.11135,0.59926 -5.98801,-0.23708 -0.0964,0.29725 -0.2395,0.46562 -0.42923,0.50513 -0.24346,-0.0284 -0.48694,-0.0569 -0.73042,-0.0854 -0.24346,-0.0284 -0.4063,0.0451 -0.48853,0.22056 -0.0142,0.12176 -0.10907,0.14152 -0.28457,0.0593 -0.22924,-0.15017 -0.47825,-0.39522 -0.74703,-0.73515 l -0.16126,-0.20395 -0.16127,-0.20395 c -1.81341,-2.4331 -3.49087,-4.97373 -5.03239,-7.6219 -1.52725,-2.76985 -2.80393,-5.57214 -3.83005,-8.40689 -1.01185,-2.9564 -1.68538,-5.90415 -2.02061,-8.84325 -0.32096,-3.06075 -0.11388,-6.15238 0.62122,-9.27491 0.35731,-0.94537 0.99761,-1.67264 1.92088,-2.18179 0.13598,-0.10745 0.30594,-0.24183 0.50986,-0.40316 0.20397,-0.1612 0.54942,-0.21338 1.03634,-0.15653 m 2.01588,9.67565 c -0.46163,0.2546 -0.80707,0.30678 -1.03634,0.15653 -0.22922,-0.15014 -0.41736,-0.38807 -0.56442,-0.71381 -0.13279,-0.44736 -0.29247,-0.92877 -0.47906,-1.44423 -0.11856,-0.5691 -0.33911,-1.05762 -0.66166,-1.46557 -0.40627,2.42056 -0.29162,4.87109 0.34395,7.35161 0.70359,2.42686 1.66486,4.76039 2.88381,7.0006 1.16524,2.17231 2.47353,4.1762 3.92489,6.01168 1.51937,1.78179 2.88457,3.29874 4.09561,4.55084 1.0498,1.04821 2.0403,1.81182 2.97151,2.29082 0.87746,0.41108 1.70985,0.41581 2.49718,0.0142 0.90906,-0.38732 1.73354,-1.37071 2.47342,-2.95016 1.06873,-2.28135 1.52877,-4.63387 1.38013,-7.05757 -0.14864,-2.42362 -0.58186,-4.78798 -1.29966,-7.0931 -0.48064,-1.16673 -1.16364,-2.44969 -2.04901,-3.84891 -0.87114,-1.52085 -1.89959,-2.75164 -3.08535,-3.69237 -1.57626,-1.29476 -3.34698,-1.71765 -5.31215,-1.26868 -1.95093,0.32734 -3.97854,1.04671 -6.08285,2.15812'\n" +
	"         style='font-size:130px;font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;text-align:start;line-height:125%;writing-mode:lr-tb;text-anchor:start;font-family:Nanum Brush Script;fill-opacity:1;fill:#003380;stroke:none;stroke-opacity:1' />\n" +
	"    </g>\n" +
	"  </g>\n" +
	"</svg>\n"

var dingoInlinePng = "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"

var editTemplate = "<!doctype html>\n" +
	"<head>\n" +
	"	<meta charset=\"utf-8\">\n" +
	"	<title>Dingo - Template Edit</title>\n" +
	//codeMirrorCSS() +
	"{{.Stylesheets}}\n" +
	"	<style>\n" +
	".clear {clear:both;}\n" +
	"html {background-color:rgb(0,25,50);}\n" +
	"body {margin:0}\n" +
	"a {text-decoration:none;color:rgb(0,85,212);}\n" +
	"a:hover {color:rgb(0,25,50);}\n" +
	"header {width:345px;margin:0 auto;}\n" +
	"header svg {float:left;}\n" +
	"header h1{float:left;margin:3px 0 3px 15px;padding:5px 0 0 0;color:rgb(220,220,235);text-shadow:1px 1px 2px rgb(0,85,212);}\n" +
	"article {margin:0 auto;background-color:white;}\n" +
	".templateEditor {width:1100px;margin:0 auto;}\n" +

	".templateEditor[multi='true'] > nav {float:left;width:250px;padding:5px;}\n" +
	".templateEditor[multi='true'] > nav > a {display:block}\n" +
	".templateEditor[multi='true'] form {float:left;width:840px;}\n" +

	".templateEditor input {display:inline;}\n" +
	".templateEditor p {display:inline;}\n" +
	".templateEditor textarea {width:100%;}\n" +
	"footer {background-color:rgb(0,25,50);border-top:solid 1px rgb(0,85,212);padding: 10px;color:rgb(235,235,245);text-align:center;}\n" +
	"footer .dVer {font-style:italic;}\n" +
	"footer a:hover {color:rgb(235,235,245);}\n" +
	".CodeMirror {border:1px solid;}\n" +
	".CodeMirror,.CodeMirror-scrollbar,.CodeMirror-scroll {height:600px;}\n" +
	"" +
	"	</style>\n" +
	"</head>\n" +
	"<body>\n" +
	"<header>\n" +
	dingoSvg +
	"	<h1>Template Editor</h1>\n" +
	"	<div class='clear'></div>\n" +
	"</header>\n" +
	"<article>\n" +
	"	<div class='templateEditor' {{if .HasViews}}multi='true'{{end}}>\n" +
	"{{if .HasViews }}" +
	"       <nav>\n" +
	"{{range $k, $v := .Views}}" +
	"           <a href='{{$.URL}}?name={{$k |urlquery}}'>{{$k |html}}</a>\n" +
	"{{end}}" +
	"       </nav>\n" +
	"{{end}}" +
	"		<form method=\"post\">\n" +
	"		    <textarea id=\"code\" name=\"content\" rows=\"35\" cols=\"120\">" + "{{printf \"%s\" .Content |html}}" + "</textarea><br>\n" +
	"		    <input type=\"submit\" value=\"Save\">\n" +
	"		    {{if not .HasViews}}<a href='{{.DoneURL}}'><button type='button'>Done</button></a>\n{{end}}" +
	"{{if .IsAction}}" +
	"{{if .WasSaved}}" +
	"           <p style='color:rgb(0,25,50)'>Saved!</p>\n" +
	"{{else}}" +
	"		    <p style='color:rgb(180,40,20)'>Error in template!</p>\n" +
	"		    <div style='font-style:italic'>{{printf \"%s\" .Error|html}}</div>" +
	"{{end}}" +
	"{{end}}" +
	"       </form>\n" +
	"	</div>\n" +
	"	<div class='clear'></div>\n" +
	"</article>\n" +
	"<footer>\n" +
	"	2013 &copy; Justin Wilson | <a href='http://juzt.in/' target='_blank'>juzt.in</a>\n" +
	"	<div class='dVer'>dingo {{.DingoVer}}</div>\n" +
	"</footer>\n" +
	//codeMirrorJS() +
	"{{.Scripts}}\n" +
	"</body>\n" +
	"</html>"
