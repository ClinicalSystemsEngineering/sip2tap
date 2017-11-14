package sipparser
import(
	"regexp"
	"fmt"
	"time"
)

func Parse(parsedmsg chan<- string,message string) {
 var pin string
 var callerid string

    //parse out the pin
	tore := regexp.MustCompile(`To:\s.*sip:(?P<pin>\d*)@`)
	matches := tore.FindStringSubmatch(message)
	if len(matches) !=0{
		pin = matches[1]
	}else{
		pin = ""
	}
	
	
	//parse out the callerid
	fromre := regexp.MustCompile(`From:\s.*"(?P<callerid>.*)".*@`)
	matches =  fromre.FindStringSubmatch(message)
	if len(matches) != 0 {
		callerid = matches[1]
	}else{
		fromre := regexp.MustCompile(`From:\ssip:(?P<callerid>.*)@`)
		matches = fromre.FindStringSubmatch(message)
		if len(matches) != 0{
			callerid = matches[1]
		}else{
			callerid = ""
		}
		
	}
	
	
	


	//check for blank pin or callerid to handle this error
	if pin == "" || callerid == "" {
		fmt.Print("SIP parser did not find a pin and callerid.\n\n")
		return
	} else {
		//print out the callerid and pin
		fmt.Printf("\n\nCaller is:<%v>\n", callerid)
		fmt.Printf("Pin is:<%v>\n", pin)
		fmt.Printf("time:%v\n\n",time.Now())
	}
	// close the sip connection.
//fmt.Print("\nPin and Callerid added to the msgqueue.\n\n")
	parsedmsg <- pin + ";" + callerid


}