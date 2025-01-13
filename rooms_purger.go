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

type deleteRoomResponse struct {
	DeleteId string `json:"delete_id"`
}

type deleteStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
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
	userRooms, err := rp.api.getUsersRooms(rp.user)
	if err != nil {
		return err
	}

	if len(userRooms) == 0 {
		fmt.Print("The user does not belong to any room. This script won't delete all rooms.")
		return nil
	}

	rooms, err := rp.api.getAllRooms()
	if err != nil {
		return err
	}

	userRoomsIndexed := make(map[string]bool, len(userRooms))
	for _, room := range userRooms {
		userRoomsIndexed[room] = true
	}

	fmt.Printf("The user belongs to %d rooms, the server has %d rooms.\n", len(userRoomsIndexed), len(rooms))

	roomsToDelete := make([]Room, 0, len(rooms)-len(userRoomsIndexed))
	for _, room := range rooms {
		if !userRoomsIndexed[room.Id] {
			roomsToDelete = append(roomsToDelete, room)
		}
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
	group.SetLimit(10)

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

	currentLine := 0
	lines := 0
	columns, _, err := term.GetSize(0)
	if err != nil {
		fmt.Print("Err: ", err)
	}

	fmt.Print("\n\n")

	lineMaxLen := columns - 10
	roomsDone := 0
	errors := map[Room]error{}

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

		fmt.Print("\033[2F")

		roomsDone = 0
		for line, roomState := range state.rooms {
			if roomState.done {
				roomsDone++
				continue
			}

			if roomState.started {
				if line < currentLine {
					fmt.Printf("\033[%dF", currentLine-line)
					currentLine = line
				} else if line > currentLine {
					if line > lines {
						// Job are played in order, so if all goes well the diff is only one
						fmt.Print("\n")
						lines++
					} else {
						fmt.Printf("\033[%dE", line-currentLine)
					}

					currentLine = line
				}

				roomDisplayName, columnsCount := getRoomDisplayName(roomState.room, lineMaxLen)
				fmt.Print(roomDisplayName)
				if roomState.finished {
					for i := columnsCount; i < columns-8; i++ {
						fmt.Print(" ")
					}

					fmt.Print(" DELETED")
					roomState.done = true
					roomsDone++
				} else if roomState.err != nil {
					for i := columnsCount; i < columns-6; i++ {
						fmt.Print(" ")
					}

					fmt.Print(" ERROR")
					roomState.done = true
					roomsDone++
					errors[roomState.room] = roomState.err
				} else {
					duration := fmt.Sprintf(" %.0fs", time.Since(roomState.startedAt).Seconds())

					for i := columnsCount; i < columns-len(duration); i++ {
						fmt.Print(" ")
					}

					fmt.Print(duration)
				}
			}
		}

		// We go to the bottom of jobs lines
		if currentLine < lines {
			fmt.Printf("\033[%dE", lines-currentLine)
			currentLine = lines
		}

		// We erase the next two lines.
		fmt.Print("\n\033[K")
		fmt.Print("\n\033[K")

		// And we write the counter
		counter := fmt.Sprintf("%d / %d (%s)", roomsDone, len(state.rooms), time.Since(start).Round(time.Second))
		b := strings.Builder{}
		for i := 0; i < columns-len(counter); i++ {
			fmt.Fprint(&b, " ")
		}

		fmt.Fprint(&b, counter)
		fmt.Print(b.String())

		time.Sleep(1 * time.Second)
	}

	fmt.Print()

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
	payload, err := rp.api.deleteRoom(roomId)
	if err != nil {
		return err
	}
	deleteId := payload.DeleteId

	for {
		payload, err := rp.api.getDeleteStatus(deleteId)
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
