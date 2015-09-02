package commands

import (
	"time"

	ui "github.com/gizak/termui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (env *Env) Stats(cmd *cobra.Command, args []string) {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	ui.UseTheme("helloworld")

	s := ui.NewPar(env.db.Scrobbles())
	s.Border.Label = "LocalFM"
	s.Height = 3

	rec := ui.NewPar(env.db.RecentTracks())
	rec.Border.Label = "Recent Tracks"
	rec.Height = (viper.GetInt("main.recent_tracks") + 2)

	topart := ui.NewPar(env.db.TopArtists())
	topart.Border.Label = "Top Artists"
	topart.Height = (viper.GetInt("main.top_artists") + 2)

	topalbs := ui.NewPar(env.db.TopAlbums())
	topalbs.Border.Label = "Top Albums"
	topalbs.Height = (viper.GetInt("main.top_albums") + 2)

	topsongs := ui.NewPar(env.db.TopSongs())
	topsongs.Border.Label = "Top Songs"
	topsongs.Height = (viper.GetInt("main.top_songs") + 2)

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, s)),
		ui.NewRow(
			ui.NewCol(12, 0, rec)),
		ui.NewRow(
			ui.NewCol(6, 0, topart),
			ui.NewCol(6, 0, topalbs)),
		ui.NewRow(
			ui.NewCol(12, 0, topsongs)))

	ui.Body.Align()

	done := make(chan bool)
	redraw := make(chan bool)

	update := func() {
		time.Sleep(time.Second / 2)
		redraw <- true
	}

	evt := ui.EventCh()

	ui.Render(ui.Body)
	go update()

	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey && e.Ch == 'q' {
				return
			}
			if e.Type == ui.EventResize {
				ui.Body.Width = ui.TermWidth()
				ui.Body.Align()
				go func() { redraw <- true }()
			}
		case <-done:
			return
		case <-redraw:
			ui.Render(ui.Body)
		}
	}
}
