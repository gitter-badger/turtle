/*
 *  Turtle - Rock Solid Cluster Management
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

/*
import (
	"time"

	ui "github.com/gizak/termui"
)

const (
	renderInterval = 300 * time.Millisecond
)

func init() {
	// Add this command.
	AddCommand("watch", new(CmdWatch))
}

type CmdWatch struct{}

func (c CmdWatch) Help() string {
	return "watch the apps state."
}

func (c CmdWatch) Run(args []string) error {
	// Initialize the UI.
	err := ui.Init()
	if err != nil {
		return err
	}
	defer ui.Close()

	// Widgets.
	data := []int{4, 2, 1, 6, 3, 9, 1, 4, 2, 15, 14, 9, 8, 6, 10, 13, 15, 12, 10, 5, 3, 6, 1, 7, 10, 10, 14, 13, 6}
	spl0 := ui.NewSparkline()
	spl0.Data = data[3:]
	spl0.Title = "Sparkline 0"
	spl0.LineColor = ui.ColorGreen

	// single
	spls0 := ui.NewSparklines(spl0)
	spls0.Height = 2
	spls0.Width = 20
	spls0.HasBorder = false

	spl1 := ui.NewSparkline()
	spl1.Data = data
	spl1.Title = "Sparkline 1"
	spl1.LineColor = ui.ColorRed

	spl2 := ui.NewSparkline()
	spl2.Data = data[5:]
	spl2.Title = "Sparkline 2"
	spl2.LineColor = ui.ColorMagenta

	// group
	spls1 := ui.NewSparklines(spl0, spl1, spl2)
	spls1.Height = 8
	spls1.Width = 20
	spls1.Y = 3
	spls1.Border.Label = "Group Sparklines"

	spl3 := ui.NewSparkline()
	spl3.Data = data
	spl3.Title = "Enlarged Sparkline"
	spl3.Height = 8
	spl3.LineColor = ui.ColorYellow

	spls2 := ui.NewSparklines(spl3)
	spls2.Height = 11
	spls2.Width = 30
	spls2.Border.FgColor = ui.ColorCyan
	spls2.X = 21
	spls2.Border.Label = "Tweeked Sparkline"

	g0 := ui.NewGauge()
	g0.Percent = 40
	g0.Width = 50
	g0.Height = 3
	g0.Border.Label = "Slim Gauge"
	g0.BarColor = ui.ColorRed
	g0.Border.FgColor = ui.ColorWhite
	g0.Border.LabelFgColor = ui.ColorCyan

	// Set the theme.
	ui.UseTheme("helloworld")

	// Create the layout.
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, spls0),
			ui.NewCol(6, 0, spls1)),
		ui.NewRow(
			ui.NewCol(6, 0, spls2),
			ui.NewCol(6, 0, g0)))

	// The quit channel.
	quit := make(chan bool)

	// Calculate layout
	ui.Body.Width = ui.TermWidth()
	ui.Body.Align()

	// Render the body.
	ui.Render(ui.Body)

	// The render timer.
	timer := time.NewTimer(renderInterval)

	// The render loop.
	evt := ui.EventCh()
	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey && e.Ch == 'q' {
				return nil
			}
			if e.Type == ui.EventResize {
				ui.Body.Width = ui.TermWidth()
				ui.Body.Align()
			}
		case <-quit:
			return nil
		case <-timer.C:
		}

		// Reset the timer.
		timer.Reset(renderInterval)

		// Render the body.
		ui.Render(ui.Body)
	}

	return nil
}
*/
