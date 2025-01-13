package synapsecleaner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SynapseAPI struct {
	accessToken string
	client      *http.Client
	server      string
}

func NewSynapseAPI(accessToken string, server string) SynapseAPI {
	client := &http.Client{}
	return SynapseAPI{
		accessToken: accessToken,
		client:      client,
		server:      server,
	}
}

func (api SynapseAPI) getUsersRooms(user string) ([]string, error) {
	fmt.Print("Fetching user's rooms... ")
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_synapse/admin/v1/users/%s/joined_rooms", api.server, user), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.sendQuery(req)
	if err != nil {
		return nil, err
	}

	var userRooms joinedRoomsResponse
	err = json.Unmarshal(resp, &userRooms)
	if err != nil {
		return nil, err
	}

	fmt.Print("OK\n")

	return userRooms.JoinedRooms, nil
}

func (api SynapseAPI) getAllRooms() ([]Room, error) {
	fmt.Print("Get all rooms in the server... ")
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_synapse/admin/v1/rooms?limit=100000", api.server), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.sendQuery(req)
	if err != nil {
		return nil, err
	}

	var roomsPayload roomsResponse
	err = json.Unmarshal(resp, &roomsPayload)
	if err != nil {
		return nil, err
	}

	rooms := []Room{}
	for _, room := range roomsPayload.Rooms {
		rooms = append(rooms, Room{
			CanonicalAlias: room.CanonicalAlias,
			Id:             room.RoomId,
			Name:           room.Name,
		})
	}

	fmt.Print("OK\n")

	return rooms, nil
}

func (api SynapseAPI) deleteRoom(roomId string) (deleteRoomResponse, error) {
	payload := deleteRoomResponse{}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/_synapse/admin/v2/rooms/%s", api.server, roomId),
		bytes.NewBuffer([]byte(`{"purge": true}`)),
	)
	if err != nil {
		return payload, err
	}

	resp, err := api.sendQuery(req)
	if err != nil {
		return payload, err
	}

	err = json.Unmarshal(resp, &payload)
	if err != nil {
		return payload, err
	}

	return payload, nil
}

func (api SynapseAPI) getDeleteStatus(deleteId string) (deleteStatusResponse, error) {
	payload := deleteStatusResponse{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_synapse/admin/v2/rooms/delete_status/%s", api.server, deleteId), nil)
	if err != nil {
		return payload, err
	}

	resp, err := api.sendQuery(req)
	if err != nil {
		return payload, err
	}

	err = json.Unmarshal(resp, &payload)
	if err != nil {
		return payload, err
	}

	return payload, nil
}

func (api SynapseAPI) sendQuery(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.accessToken))
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("response status code is %d: %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}
