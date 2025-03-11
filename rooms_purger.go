package synapsecleaner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"
)

type joinedRoomsResponse struct {
	JoinedRooms []string `json:"joined_rooms"`
	Total       int
}

type roomsResponseRoom struct {
	CanonicalAlias string `json:"canonical_alias"`
	Name           string `json:"name"`
	RoomId         string `json:"room_id"`
}

type roomsResponse struct {
	Rooms []roomsResponseRoom `json:"rooms"`
}

type RoomsPurger struct {
	user string
	api  SynapseAPI
}

func NewRoomsPurger(accessToken string, user string, server string) RoomsPurger {
	return RoomsPurger{
		api:  NewSynapseAPI(accessToken, server),
		user: user,
	}
}

func (rp RoomsPurger) Do() error {
	userRooms, err := rp.api.GetUsersRooms(rp.user)
	if err != nil {
		return err
	}

	if len(userRooms) == 0 {
		fmt.Print("The user does not belong to any room. This script won't delete all rooms.")
		return nil
	}

	rooms, err := rp.api.GetAllRooms()
	if err != nil {
		return err
	}

	fmt.Printf("The user belongs to %d rooms, the server has %d rooms\n", len(userRooms), len(rooms))
	roomsToDelete := DiffSlicesFunc(
		rooms,
		userRooms,
		func(r Room) string { return r.Id },
		func(roomId string) string { return roomId },
	)

	if len(roomsToDelete) == 0 {
		fmt.Print("No rooms to delete.")
		os.Exit(0)
	}

	fmt.Printf("\nYou are about to *PERMANENTLY* delete %d rooms, do you want to proceed?\nAnything else than \"yes\" will stop the process: ", len(roomsToDelete))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Print()

	if input != "yes\n" {
		fmt.Print("Stopping there.")
		return nil
	}

	err = rp.deleteRooms(roomsToDelete)
	if err != nil {
		return err
	}

	fmt.Print("Task done. If you are using PostgreSQL you probably want to run `vacuum full` on the Synapse database to return the free space to the OS.")
	return nil
}

type DeletionScreenRoomState struct {
	started   bool
	err       error
	finished  bool
	index     int
	done      bool
	room      Room
	startedAt time.Time
}

type DeletionScreenState struct {
	rooms []*DeletionScreenRoomState
}

type RoomLine struct {
	id      string
	content string
}

func (rl RoomLine) Id() string {
	return rl.id
}

func (rl RoomLine) String() string {
	return rl.content
}

func (rp RoomsPurger) deleteRooms(roomsToDelete []Room) error {
	fmt.Printf("Deleting %d rooms...\n\n", len(roomsToDelete))
	start := time.Now()

	state := DeletionScreenState{}
	for idx, room := range roomsToDelete {
		state.rooms = append(
			state.rooms,
			&DeletionScreenRoomState{
				started:  false,
				err:      nil,
				finished: false,
				index:    idx,
				done:     false,
				room:     room,
			},
		)
	}

	group := errgroup.Group{}
	group.SetLimit(20)

	go func() {
		for _, roomState := range state.rooms {
			group.Go(
				func() error {
					roomState.started = true
					roomState.startedAt = time.Now()

					err := rp.deleteRoom(roomState.room.Id)
					if err != nil {
						roomState.err = err
					} else {
						roomState.finished = true
					}

					return nil
				},
			)
		}
	}()

	jobsDone := 0
	errors := map[Room]error{}

	columns, _, err := term.GetSize(0)
	if err != nil {
		fmt.Print("Err: ", err)
	}

	lineMaxLen := columns - 10

	printer, err := NewLinesPrinterWithFooter(2)
	if err != nil {
		return err
	}

	for {
		done := true
		for _, roomState := range state.rooms {
			if !roomState.done {
				done = false
				break
			}
		}

		if done {
			break
		}

		for _, roomState := range state.rooms {
			if roomState.done {
				continue
			}

			if roomState.started {
				roomDisplayName, columnsCount := getRoomDisplayName(roomState.room, lineMaxLen)
				b := strings.Builder{}
				fmt.Fprint(&b, roomDisplayName)
				if roomState.finished {
					for i := columnsCount; i < columns-8; i++ {
						fmt.Fprint(&b, " ")
					}

					fmt.Fprint(&b, " DELETED")
					roomState.done = true
					jobsDone++
				} else if roomState.err != nil {
					for i := columnsCount; i < columns-6; i++ {
						fmt.Fprint(&b, " ")
					}

					fmt.Fprint(&b, " ERROR")
					roomState.done = true
					jobsDone++
					errors[roomState.room] = roomState.err
				} else {
					duration := fmt.Sprintf(" %s", time.Since(roomState.startedAt).Round(time.Second))

					for i := columnsCount; i < columns-len(duration); i++ {
						fmt.Fprint(&b, " ")
					}

					fmt.Fprint(&b, duration)
				}

				printer.Print(RoomLine{id: roomState.room.Id, content: b.String()})
			}
		}

		// We write the counter
		counter := fmt.Sprintf("%d / %d (%s)", jobsDone, len(state.rooms), time.Since(start).Round(time.Second))
		b := strings.Builder{}
		fmt.Fprintln(&b)
		fmt.Fprint(&b, strings.Repeat(" ", columns-len(counter)))
		fmt.Fprint(&b, counter)
		printer.PrintFooter(b.String())

		time.Sleep(1 * time.Second)
	}

	printer.Exit()
	fmt.Println()

	if len(errors) > 0 {
		fmt.Printf("There was %d errors during the process:\n\n", len(errors))
		for room, err := range errors {
			roomDisplayName, _ := getRoomDisplayName(room, columns-20)
			fmt.Print(roomDisplayName, ":\n")
			fmt.Print(err, "\n\n")
		}
	}

	fmt.Print("\n")

	return nil
}

func getRoomDisplayName(room Room, width int) (string, int) {
	b := strings.Builder{}
	columnsCount := 0
	if room.Name != "" {
		name := room.Name[:min(len(room.Name), width)]
		fmt.Fprint(&b, name)
		columnsCount += runewidth.StringWidth(name)

		if room.CanonicalAlias != "" {
			alias := room.CanonicalAlias[:min(len(room.CanonicalAlias), width-b.Len()-2)]
			fmt.Fprint(&b, " (", alias, ")")
			columnsCount += runewidth.StringWidth(alias) + 3
		}
	} else if room.CanonicalAlias != "" {
		alias := room.CanonicalAlias[:min(len(room.CanonicalAlias), width)]
		fmt.Fprint(&b, alias)
		columnsCount += runewidth.StringWidth(alias)
	} else {
		fmt.Fprint(&b, room.Id)
		columnsCount += len(room.Id)
	}

	return b.String(), columnsCount
}

func (rp RoomsPurger) deleteRoom(roomId string) error {
	payload, err := rp.api.DeleteRoom(roomId)
	if err != nil {
		return err
	}
	deleteId := payload.DeleteId

	for {
		payload, err := rp.api.GetDeleteStatus(deleteId)
		if err != nil {
			return err
		}

		if payload.Status == "complete" {
			return nil
		} else if payload.Status == "failed" {
			return fmt.Errorf("error from the homeserver: %s", payload.Error)
		}

		time.Sleep(2 * time.Second)
	}
}
