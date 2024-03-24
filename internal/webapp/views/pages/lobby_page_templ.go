// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.635
package pages

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "kingscomp/internal/webapp/views/layout"

func LobbyPage(lobbyId string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
			if !templ_7745c5c3_IsBuffer {
				templ_7745c5c3_Buffer = templ.GetBuffer()
				defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"center\" x-init=\"\n    $store.lobby.setLobbyId($root.dataset.lobbyid)\n    await $store.lobby.initLobby()\n\" x-data=\"\" data-lobbyid=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(lobbyId)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/webapp/views/pages/lobby_page.templ`, Line: 12, Col: 33}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><template x-if=\"$store.lobby.isInit\"><div style=\"width: 90%;margin-left: 5%\"><template x-if=\"[&#39;created&#39;,&#39;get-ready&#39;].includes($store.lobby.currentLobby.state)\"><div class=\"anim-fade-in\" style=\"width: 100%\"><h2>خوش آمدید <span x-text=\"$store.lobby.currentPlayer.display_name\"></span></h2><p style=\"margin-bottom: 20px\">درحال انتظار برای متصل شدن بقیه بازیکنان</p><template x-for=\"(value, index) in $store.lobby.currentLobby.participants\"><div class=\"box-with-border flex-row\" style=\"margin-bottom: 4px;\"><div x-text=\"value.displayName\"></div><div x-show=\"value.isReady\" class=\"hint\"><i class=\"gg-check-r\"></i> <span>متصل شده</span></div><div x-show=\"!value.isReady &amp;&amp; !value.isResigned\" class=\"hint\"><i class=\"gg-loadbar\"></i> انتظار اتصال</div><div x-show=\"value.isResigned\" class=\"hint\"><i class=\"gg-smile-sad\"></i> انصراف داده</div></div></template><div style=\"padding-top: 10px\"><p class=\"hint\" x-show=\"$store.lobby.currentLobby.state === &#39;created&#39;\" x-transition>به محض اتصال همه\u200cی بازیکنان بازی شروع میشود</p><p class=\"hint\" x-transition x-show=\"$store.lobby.currentLobby.state === &#39;get-ready&#39;\"><i class=\"gg-check-r\"></i> همه وصل شدن و بازی داره شروع میشه!</p></div></div></template><template x-if=\"[&#39;started&#39;].includes($store.lobby.currentLobby.state)\"><div class=\"anim-fade-in center\" style=\"width: 100%\"><p class=\"hint\">شماره سوال: <span x-text=\"$store.lobby.currentLobby.gameInfo.currentQuestion.index+1\"></span></p><div class=\"time-indicator-holder\" style=\"width: 100%\"><div class=\"time-indicator\" :style=\"&#39;margin-left: calc( &#39;+$store.lobby.timerPercent+&#39;% - 80px)&#39;\"></div></div><h2 style=\"width: 100%; margin-top: 5px;\" x-text=\"$store.lobby.currentLobby.gameInfo.currentQuestion.question\"></h2><div style=\"padding-top:30px; width: 100%;display: flex;flex-wrap: wrap;justify-content: space-between\"><template x-for=\"(value,index) in $store.lobby.currentLobby.gameInfo.currentQuestion.choices\"><button :class=\"&#39;tg-button &#39;+($store.lobby.currentLobby.gameInfo.currentQuestion.index &lt;= $store.lobby.lastQuestionAnswered ? &#39;answered&#39;:&#39;&#39;)\" style=\"width: 45%;margin-bottom: 10px\" @click.prevent=\"$store.lobby.answered(index)\"><span x-text=\"value\"></span></button></template></div><div class=\"active-users\" style=\"width: 100%\"><template x-for=\"participant in $store.lobby.currentLobby.participants\"><div style=\"overflow: hidden;margin-top: 5px\" class=\"box-with-border flex-row\"><strong x-text=\"participant.displayName\"></strong><template x-if=\"participant.isResigned\"><span style=\"font-size: 12px\" class=\"hint\">این کاربر شکست خورده!</span></template><template x-if=\"!participant.isResigned\"><div style=\"display: flex;justify-content: flex-start;align-items: center\"><template x-if=\"!!participant.history.answerHistory\" x-for=\"answer in participant.history.answerHistory\"><p style=\"padding-left: 4px;\"><i x-show=\"answer.correct\" class=\"gg-check-o\"></i> <i x-show=\"!answer.correct\" class=\"gg-radio-check\"></i></p></template></div></template></div></template></div></div></template><template x-if=\"[&#39;ended&#39;].includes($store.lobby.currentLobby.state)\"><div class=\"anim-fade-in\" style=\"width: 100%\"><div class=\"center\"><h4 style=\"padding: 30px\">بازی به اتمام رسید!</h4><p class=\"hint\">برنده بازی</p><h1 class=\"flex-row\"><i class=\"gg-crown\"></i> <strong x-text=\"$store.lobby.currentLobby.result.winner\"></strong></h1><h4 class=\"hint\">رتبه اول را با <strong x-text=\"$store.lobby.currentLobby.result.winnerScore\"></strong> امتیاز کسب کرد</h4><div style=\"width: 100%\"><div class=\"active-users\" style=\"width: 100%\"><template x-for=\"(item, index) in $store.lobby.currentLobby.result.leaderboard\"><div style=\"overflow: hidden;margin-top: 5px\" class=\"box-with-border flex-row\"><p><span x-text=\"&#39;#&#39;+(index+1)\"></span> <strong x-text=\"item.displayName\"></strong></p><strong x-text=\"item.score\"></strong></div></template></div><em class=\"hint\" style=\"font-size: 12px\">سرعت پاسخدهی در امتیاز شما نیز تاثیر میگذارد</em><div style=\"margin-top: 20px\"><button class=\"tg-button-bordered\" style=\"width: 100%\" @click.prevent=\"WebApp.close()\">بستن بازی</button></div></div></div></div></template></div></template><template x-if=\"!$store.lobby.isInit\"><div class=\"center\"><div><div class=\"spinner\"></div></div><p class=\"hint\">درحال دریافت اطلاعات بازی</p></div></template></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !templ_7745c5c3_IsBuffer {
				_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = layout.Base().Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
