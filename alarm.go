package asdf

import (
	. "fmt"
	"time"
)

func Timer_per_day(hour int, mint int) <-chan time.Time {
	ch := make(chan time.Time)
	go run_per_day(ch, hour, mint)
	return ch
}

func Timer_per_month(day int) <-chan time.Time {
	ch := make(chan time.Time)
	go run_per_month(ch, day)
	return ch
}

func Timer_per_week(weekday int) <-chan time.Time {
	ch := make(chan time.Time)
	go run_per_week(ch, weekday)
	return ch
}

func run_per_week(ch chan time.Time, wk int) {

	var dur time.Duration = 1 * time.Hour

	day := time.Now().Weekday()

	if day == time.Weekday(wk) {
		dur = 24 * time.Hour
	}

	for {
		day = time.Now().Weekday()

		if day != time.Weekday(wk) {
			time.Sleep(dur)
			dur = 1 * time.Hour
		} else {
			ch <- time.Now()
			return
		}
	}
}

func run_per_day(ch chan time.Time, h int, m int) {

	var dur time.Duration = 1 * time.Minute

	for {
		hour := time.Now().Hour()
		minute := time.Now().Minute()

		if hour == h && minute == m {
			ch <- time.Now()
			return
		} else {
			time.Sleep(dur)
		}
	}
}

func run_per_month(ch chan time.Time, d int) {

	var dur time.Duration = 1 * time.Hour

	_, _, day := time.Now().Date()

	if day == d {
		dur = 24 * time.Hour
	}

	for {
		_, _, day = time.Now().Date()

		if day != d {
			time.Sleep(dur)
			dur = 1 * time.Hour
		} else {
			ch <- time.Now()
			return
		}
	}
}

/*
func main() {
	//ch := timer_per_month(1)
	Println(time.Now())
	//ch := timer_per_month(1)
	ch := timer_per_week(1)
	t := <-ch
	Println("Now time is", t)
	Println("True time is", time.Now())
}
*/
