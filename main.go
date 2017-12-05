package main

import(
  "net"
  "net/http"
  "fmt"
  "bufio"
  "log"
  "io/ioutil"
  "encoding/json"
  "strconv"
  "strings"
)

type ForexData struct {
  Base  string `json:"base"`
  Date  string `json:"date"`
  Rates map[string]float64 `json:"rates"`
}

type Rate struct {
  Rate    float64
  Result  float64
}

func HandleBot(conn net.Conn) {
  defer conn.Close()
  fmt.Println("New connection received")  
  c := make(chan Rate)
  input := make(map[string]string)

  fmt.Fprintln(conn, "BOT: Hi, this is forex bot.. May I know what is the currency that you want to convert?")
  fmt.Fprint(conn, "- ")

  scanner := bufio.NewScanner(conn)
  for {
    scanner.Scan()
    input["base"] = strings.ToUpper(scanner.Text())
    fmt.Fprintln(conn, "BOT: How much do you want to convert", input["base"])
    fmt.Fprint(conn, "- ")

    scanner.Scan()
    input["amount"] = scanner.Text()
    fmt.Fprintln(conn, "BOT: To which currency do you want to convert ")
    fmt.Fprint(conn, "- ")

    scanner.Scan()
    input["sym"] = strings.ToUpper(scanner.Text())
    fmt.Fprintln(conn, "BOT: Calculating...")
    go CalculateRate(conn, input, c)
    go PrintResult(conn, input, c)

    fmt.Fprintln(conn, "BOT: Meanwhile, anything else do you want to convert?")
    fmt.Fprint(conn, "- ")
  }
}

func ConsumeForex(conn net.Conn, base string, sym string) []byte {
  li := "https://api.fixer.io/latest?base=" + base +"&symbols=" + sym
  fmt.Println("Calling..", li)
  response, err := http.Get(li)
  if err != nil {
      fmt.Fprintf(conn, "The HTTP request failed with error %s\n", err)
  }
  data, _ := ioutil.ReadAll(response.Body)
  fmt.Println(string(data))
  
  return data
}

func NewForex(bs []byte) ForexData {
  var f ForexData
  err := json.Unmarshal(bs, &f)
  if err != nil {
    fmt.Println(err)
  }
  
  return f
}

func CalculateRate(conn net.Conn, input map[string]string, c chan Rate){
  bs := ConsumeForex(conn, input["base"], input["sym"])
  fd := NewForex(bs)
  famount, _ := strconv.ParseFloat(input["amount"], 64)
  c <- Rate{ fd.Rates[input["sym"]], fd.Rates[input["sym"]] * famount }
}

func PrintResult(conn net.Conn, input map[string]string, c chan Rate){
  res := <-c
  fmt.Println(res)
  ra := strconv.FormatFloat(res.Rate, 'f', 2, 64)
  re := strconv.FormatFloat(res.Result, 'f', 2, 64)
  fmt.Fprintf(conn, "\nBOT: Thank you for waiting, the rate for %s is %s therefore you will get %s %s\n", input["base"], ra, input["sym"], re)
  fmt.Fprint(conn, "- ")
}

func main() {
  li, err := net.Listen("tcp", ":8080")
  if err != nil {
    log.Panic(err)
  }
  defer li.Close()

  for {
    conn, err := li.Accept()
    if err != nil {
      fmt.Println(err)
    }
    go HandleBot(conn)
  }
}