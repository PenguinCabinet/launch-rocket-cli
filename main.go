package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

type rocket_AA_t struct {
	raw_aa          string
	split_aa        [][]rune
	width           int
	height          int
	bottom_offset_y int
}

func new_rocket_AA_t(aa string, bottom_offset_y int) rocket_AA_t {
	temp := strings.Split(aa, "\n")
	var temp2 [][]rune
	for _, v := range temp {
		temp2 = append(temp2, []rune(v))
	}

	result := rocket_AA_t{
		raw_aa:          aa,
		split_aa:        temp2,
		width:           -1,
		height:          strings.Count(aa, "\n") + 1,
		bottom_offset_y: bottom_offset_y,
	}

	for _, v := range result.split_aa {
		result.width = max(result.width, len(v))
	}

	return result
}

func term_GetSize() (int, int) {
	term_width, term_height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Error getting terminal size: %s\n", err)
	}
	return term_width, term_height
}

func launch_rocket_anime_render(rocket_AA rocket_AA_t, waiting_time_before_start time.Duration, waiting_time_before_end time.Duration, init_rocket_x int, init_rocket_y int, rocket_dy func(int, int) int, rocket_dx func(int, int) int, acceleration float64, time_offset int) {

	rocket_x := init_rocket_x
	rocket_y := init_rocket_y

	term_width, term_height := term_GetSize()
	for t := 0; rocket_y >= -rocket_AA.height && rocket_y <= term_height; t++ {
		term_width, term_height = term_GetSize()
		fmt.Print("\033[H\033[2J")
		for y := rocket_AA.height - 1; y >= 0; y-- {
			for x := 0; x < len(rocket_AA.split_aa[y]); x++ {
				if 0 <= rocket_x+x && rocket_x+x < term_width && 0 <= rocket_y+y && rocket_y+y < term_height {
					fmt.Printf("\x1b[%d;%dH", rocket_y+y, rocket_x+x)
					fmt.Printf("%c", rocket_AA.split_aa[y][x])
				}
			}
		}
		rocket_y = rocket_dy(rocket_y, t)
		rocket_x = rocket_dx(rocket_x, t)
		if t == 0 {
			time.Sleep(waiting_time_before_start)
		}

		//time.Sleep(time.Millisecond * 300)

		/*現実世界の物理法則に合わせようとしたが、あまり良い演出にはならなかった*/
		//time.Sleep(time.Millisecond * time.Duration(math.Sqrt(2*float64(rocket_x)/acceleration)-math.Sqrt(2*float64(rocket_x-1)/acceleration)))

		time.Sleep(time.Millisecond * time.Duration(1/(acceleration*float64(t+time_offset))))
	}

	time.Sleep(waiting_time_before_end)

}

func main() {

	rocket_AA := new_rocket_AA_t(`     /\ 
    /  \
   /++++\
  /  ()  \
  |      |
  |  []  |
  |      |
  |      |
 /|______|\
   /_|_|_\
     / \
    /___\`, 3)

	rocket_AA_fall := new_rocket_AA_t(` \|______|/
  |      |
  |      |
  |  []  |
  |      |
  \  ()  /
   \++++/
    \  /
     \/`, 0)

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "fall",
				Aliases: []string{
					"f",
				},
				Value: false,
				Usage: "Will the rocket crash after launch?",
			},
			&cli.Uint32Flag{
				Name: "waiting_time_before_start",
				Aliases: []string{
					"wtbs",
				},
				Value: 1000,
				Usage: "Waiting time before launch or crash begins(ms)",
			},
			&cli.Uint32Flag{
				Name: "waiting_time_before_end",
				Aliases: []string{
					"wtbe",
				},
				Value: 1000,
				Usage: "Waiting time after completing launch or crash(ms)",
			},
			&cli.Float64Flag{
				Name: "acceleration",
				Aliases: []string{
					"a",
				},
				Value: 0.001,
				Usage: "Acceleration parameters",
			},
			&cli.Float64Flag{
				Name: "fall_frequency",
				Aliases: []string{
					"ff",
				},
				Value: 0.1,
				Usage: "Frequency parameters that sway from side to side during a fall",
			},
			&cli.Float64Flag{
				Name: "fall_width",
				Aliases: []string{
					"fw",
				},
				Value: 5,
				Usage: "Wave height parameter that sways from side to side during descent",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {

			term_width, term_height := term_GetSize()

			launch_rocket_anime_render(
				rocket_AA,
				time.Duration(cmd.Uint32("waiting_time_before_start"))*time.Millisecond,
				time.Duration(cmd.Uint32("waiting_time_before_end"))*time.Millisecond,
				term_width/2,
				term_height-rocket_AA.height+rocket_AA.bottom_offset_y,
				func(y, t int) int { return y - 1 },
				func(x, t int) int { return x },
				cmd.Float64("acceleration"),
				0,
			)

			if cmd.Bool("fall") {
				term_width, term_height = term_GetSize()
				launch_rocket_anime_render(
					rocket_AA_fall,
					time.Duration(cmd.Uint32("waiting_time_before_start"))*time.Millisecond,
					time.Duration(cmd.Uint32("waiting_time_before_end"))*time.Millisecond,
					term_width/2+10,
					-rocket_AA_fall.height,
					func(y, t int) int { return y + 1 },
					func(x, t int) int {
						return term_width/2 + 10 + int(math.Round(cmd.Float64("fall_width")*math.Sin(cmd.Float64("fall_frequency")*(math.Pi/2)*float64(t))))
					},
					cmd.Float64("acceleration"),
					10,
				)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
