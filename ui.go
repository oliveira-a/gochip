// ui of the emulator goes here. This includes the root
// container, the list widget that allows the user to
// selecct a game and the game window itself.

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type listItem struct {
	// used for the display name
	name string

	// holds the path to the rom in the our embedded
	// directory. Later used in the callback function
	// defined in the client code.
	path string
}

type sidelist struct {
	container  *widget.Container
	listWidget *widget.List
}

//go:embed static/press-start-2p.ttf
var font []byte

// The side list that allows the user to select a game
func newSidelist(
	items []any,
	entrySelectedEventHandler func(args *widget.ListEntrySelectedEventArgs),
	w, h int,
) *sidelist {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  0,
				Right: 0,
			}),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			widget.GridLayoutOpts.Spacing(0, 0),
		)))

	b, _ := loadButtonImage()
	f, _ := loadFont(15, font)

	lw := widget.NewList(
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(w, h),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchVertical:    false,
				Padding:            widget.NewInsetsSimple(50),
			}),
		)),

		widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),

		// set the entries
		widget.ListOpts.Entries(items),

		widget.ListOpts.ScrollContainerOpts(
			// Set the background images/color for the list
			widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
				Idle:     image.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
				Disabled: image.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
				Mask:     image.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
			}),
		),

		widget.ListOpts.SliderOpts(
			// Set the background images/color for the background of the slider track
			widget.SliderOpts.Images(&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
			}, b),
			widget.SliderOpts.MinHandleSize(5),
			// Set how wide the track should be
			widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2))),

		// Hide the horizontal slider
		widget.ListOpts.HideHorizontalSlider(),

		// Set the font for the list options
		widget.ListOpts.EntryFontFace(f),

		// Set the label position
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),

		// Set the colors for the list
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Selected:                   color.NRGBA{R: 0, G: 255, B: 0, A: 255},     // Foreground color for the unfocused selected entry
			Unselected:                 color.NRGBA{R: 254, G: 255, B: 255, A: 255}, // Foreground color for the unfocused unselected entry
			SelectedBackground:         color.NRGBA{R: 130, G: 130, B: 200, A: 255}, // Background color for the unfocused selected entry
			SelectingBackground:        color.NRGBA{R: 130, G: 130, B: 130, A: 255}, // Background color for the unfocused being selected entry
			SelectingFocusedBackground: color.NRGBA{R: 130, G: 140, B: 170, A: 255}, // Background color for the focused being selected entry
			SelectedFocusedBackground:  color.NRGBA{R: 130, G: 130, B: 170, A: 255}, // Background color for the focused selected entry
			FocusedBackground:          color.NRGBA{R: 170, G: 170, B: 180, A: 255}, // Background color for the focused unselected entry
			DisabledUnselected:         color.NRGBA{R: 100, G: 100, B: 100, A: 255}, // Foreground color for the disabled unselected entry
			DisabledSelected:           color.NRGBA{R: 100, G: 100, B: 100, A: 255}, // Foreground color for the disabled selected entry
			DisabledSelectedBackground: color.NRGBA{R: 100, G: 100, B: 100, A: 255}, // Background color for the disabled selected entry
		}),

		// Select the property from the list item struct (if
		// any) with this callback function.
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(listItem).name
		}),

		// Provide the function to run when a list item is
		// selected.
		widget.ListOpts.EntrySelectedHandler(entrySelectedEventHandler),
	)

	root.AddChild(lw)

	sl := &sidelist{
		container:  root,
		listWidget: lw,
	}

	return sl
}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	hover := image.NewNineSliceColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	pressed := image.NewNineSliceColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func loadContextMenuButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	hover := image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	pressed := image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func loadFont(size float64, font []byte) (text.Face, error) {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(font))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil
}

func newTickRateContextMenu() *widget.Container {
	contextMenu := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
	)
	buttonImage, _ := loadContextMenuButtonImage()
	face, _ := loadFont(10, font)

	opt60TPS := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("60  TPS", face, &widget.ButtonTextColor{
			Idle:  color.NRGBA{0, 0, 0, 255},
			Hover: color.NRGBA{255, 255, 255, 255},
		}),

		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ebiten.SetTPS(60)
		}),
	)

	opt120TPS := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("120 TPS", face, &widget.ButtonTextColor{
			Idle:  color.NRGBA{0, 0, 0, 255},
			Hover: color.NRGBA{255, 255, 255, 255},
		}),

		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ebiten.SetTPS(120)
		}),
	)

	opt240TPS := widget.NewButton(
		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("240 TPS", face, &widget.ButtonTextColor{
			Idle:  color.NRGBA{0, 0, 0, 255},
			Hover: color.NRGBA{255, 255, 255, 255},
		}),

		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ebiten.SetTPS(240)
		}),
	)

	contextMenu.AddChild(opt60TPS)
	contextMenu.AddChild(opt120TPS)
	contextMenu.AddChild(opt240TPS)

	return contextMenu
}

func newSlider() *widget.Container {
	// construct a slider
	slider := widget.NewSlider(
		// Set the slider orientation - n/s vs e/w
		widget.SliderOpts.Direction(widget.DirectionHorizontal),

		// Set the minimum and maximum value for the slider
		widget.SliderOpts.MinMax(60, 240),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),

			// Set the widget's dimensions
			widget.WidgetOpts.MinSize(150, 0),
		),

		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Hover:   image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Pressed: image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
			},
		),

		// Set the size of the handle
		widget.SliderOpts.FixedHandleSize(6),

		// Set the offset to display the track
		widget.SliderOpts.TrackOffset(0),

		// Set the size to move the handle
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),

		// Set the callback to call when the slider value is changed
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			fmt.Println(args.Current, "dragging", args.Dragging)
		}),
	)

	container := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	container.AddChild(slider)

	return container
}
