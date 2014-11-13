package eve_test

import (
	"fmt"

	"github.com/nightexcessive/eve"
)

const data = `
<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <currentTime>2006-01-02 15:04:05</currentTime>
  <result>
    <rowset name="callGroups" key="groupID" columns="groupID,name,description">
      <row groupID="1" name="Account and Market" description="Market Orders, account balance and journal history." />
      <row groupID="2" name="Science and Industry" description="Datacore production and job listing." />
      <row groupID="3" name="Private Information" description="Personal information about the owner. Asset lists, skill training for characters, Private Calendar and more." />
      <row groupID="4" name="Public Information" description="Achievements such as Medals, Kill Mails, Fational Warfare Statistics and NPC Standings." />
      <row groupID="5" name="Corporation Members" description="Member information for Corporations." />
      <row groupID="6" name="Outposts and Starbases" description="Outpost and Starbase information for Corporations" />
      <row groupID="7" name="Communications" description="Private communications such as contact lists, Eve Mail and Notifications." />
    </rowset
  </result>
  <cachedUntil>2020-01-02 15:04:05</cachedUntil>
</eveapi>`

type CallGroup struct {
	ID          int    `xml:"groupID,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,attr"`
}

type OutputStruct struct {
	CallGroups []*CallGroup `eve:"callGroups"`
}

func ExampleDecode() {
	output := new(OutputStruct)
	_, err := eve.Decode([]byte(data), output)
	if err != nil {
		panic(err)
	}

	fmt.Println("Call groups:")
	for _, group := range output.CallGroups {
		fmt.Printf("%d. %22s: %s\n", group.ID, group.Name, group.Description)
	}
}
