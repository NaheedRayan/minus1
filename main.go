package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NaheedRayan/minus1/script"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var g_ctx context.Context
var g_client genai.Client
var g_cs genai.ChatSession

var g_file os.File
var g_logger *log.Logger

var g_file_2 os.File
var g_logger_2 *log.Logger


type Config struct {
	APIKey    string `json:"apikey"`
	MaxRetry int    `json:"max_retry"`
	ModelName string `json:"modelName"`
}
var config Config


func main() {


	//////////////////////////////////////////config file setup ///////////////////////////////////////////////////

	configFile, err := os.Open("config.json")
	if err != nil {
			fmt.Println("Error opening config file:", err)
			return
	}
	defer configFile.Close()

	
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
			fmt.Println("Error decoding config file:", err)
			return 

	}

	///////////////////////////////////////////////// setting up log files/////////////////////////////////////////

	file, err := os.OpenFile("cmdList.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	file.Truncate(0)//clearing the log file on load
	g_file = *file

	// custom logger with no prefix and no flags
	logger := log.New(os.Stdout , "" , 0)
	logger.SetOutput(&g_file)
	g_logger = logger


	// logger for cmds
	file_2, err := os.OpenFile("cmdLog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	file_2.Truncate(0)//clearing the log file on load
	g_file_2 = *file_2

	// custom logger with no prefix and no flags
	logger2 := log.New(os.Stdout , "" , 0)
	logger2.SetOutput(&g_file_2)
	g_logger_2 = logger2





	// Gemini setup

	// ///////////////////////////////////////////////////Context1 //////////////////////////////////////////////////////////////////
	g_ctx = context.Background()

	client, err := genai.NewClient(g_ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		log.Fatal(err)
	}
	g_client = *client

	// closing the session 1
	defer client.Close()

	system_prompt := `

	Minus1 is a Linux terminal assistant capable of executing two primary types of tasks:

	1. **General Question Answering (Task 1)**: Minus1 provides answers to general questions based on the information available. When handling Task 1, Minus1 does not require command execution, and the "taskOutput" will contain the answer. The "taskStatus" should be set to "Completed" once the answer is provided.

	2. **Command Execution (Task 2)**: Minus1 can execute commands in the Linux terminal, such as navigating the file system, creating files or directories, searching for files, and more. For Task 2:
	   - Minus1 receives a list of commands (cmdList) in the format "cmd cmdArgs".
	   - Minus1 can use commands like "ls", "pwd", "cd" , "~" , "mkdir", "find", etc., to interact with the file system.
	   - The "taskStatus" should initially be set to "Running" during command execution and updated to "Completed" or "Failed" based on the outcome.
	   - The "retryCnt" field keeps track of the number of retries if command execution fails. Minus1 may adjust the "cmdList" and attempt re-execution if necessary.



	Use this JSON schema:
	{
	"type": "object",
	"properties": {
	  "task": {
		"type": "integer",
		"enum": [1, 2],
		"description": "1 for general questions, 2 for command tasks"
	  },
	  "taskStatus": {
		"type": "string",
		"enum": ["Completed", "Running", "Failed"],
		"description": "Current status of the task"
	  },
	  "taskOutput": {
		"type": "string",
		"description": "The output or answer to the task"
	  },
	  "cmdList": {
		"type": "array",
		"items": {
		  "type": "string"
		},
		"description": "List of commands to be executed, applicable only for task 2"
	  },
	  "retryCnt": {
		"type": "integer",
		"description": "Count of the number of retries if the cmdList fails and needs to be updated"
	  }
	},
	"required": ["task", "taskStatus", "taskOutput", "cmdList" , "retryCnt" ]
		}


	`
	model := client.GenerativeModel(config.ModelName)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(system_prompt)},
	}
	// Set the `ResponseMIMEType` to output JSON
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	model.SafetySettings = []*genai.SafetySetting{
		{
		  Category:  genai.HarmCategoryHarassment,
		  Threshold: genai.HarmBlockNone,
		},
		{
		  Category:  genai.HarmCategoryHateSpeech,
		  Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	  }

	// setting global chat session 1
	cs := model.StartChat()
	g_cs = *cs

	
	///////////////////////////////////////////////Starting the new program///////////////////////////////////////////////////////////

	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	
}


// responsetype structure
type (
	errMsg error
	aiMsg  struct {
		response string
	}
	cmdExecutorMsg  struct {
		response string
	}
)


// initial model
type model struct {
	userInput   string
	err         error


	spinner    spinner.Model
	workingMsg string
	workingView bool

	retryCnt int

	viewport    viewport.Model
	textarea    textarea.Model
	messages    []string
	senderStyle lipgloss.Style
}

// color list reference
// red > green > yellow > blue > pink > cyan > grey > dark grey
//  1  >   2   >   3    >  4   >   5  >   6  >  7   >   8



// var for rendering colors
var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
var redColor = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render
var greenColor = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render
var yellowColor = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render
var cyanColor = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render
var pinkColor = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render


// extra view for help and functunality decrpition
func (e model) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • esc: Quit\n")
}


// initial model
func initialModel() model {
	ta := textarea.New()
	// ta.Placeholder = "Enter your command or question here..🗿.✅❎❌🔁✔✘🟢🚀"
	ta.Placeholder = "Enter your command or question here..🗿..lemme take you to moon 🚀"
	ta.Focus()

	ta.Prompt = "┃ "
	// ta.CharLimit = 900

	ta.SetWidth(80)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)



	vp := viewport.New(80, 10)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)


	vp.SetContent(cyanColor("Minus1 Terminal"))
	


	// spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		userInput:   "",

		spinner:    s,
		workingMsg: greenColor("Analyzing...⏳"),
		workingView: false,

		retryCnt: 0,

		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}


// init function. The blink and the spinner tick should run first
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, spinner.Tick)
}

// update function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			// Get user input
			m.userInput = m.textarea.Value()

			// Display user input immediately
			m.messages = append(m.messages, m.senderStyle.Render("You : ")+m.userInput)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			// loading spinner
			m.workingView = true
			g_logger.Println("----------------------------------------------")


			// Return a command to process AI interaction 
			return m, askAI(m.userInput)

		case tea.KeyBackspace:
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}

		default:
			m.userInput += msg.String()
		}


	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width-2
		// m.viewport.Height = msg.Height-7

		m.textarea.MaxWidth = msg.Width-2
		// m.textarea.MaxWidth = msg.Width-2
	case aiMsg:
		// Handle the AI response
	
		// Task represents the structure defined by the JSON schema.
		type Task struct {
			Task        int      `json:"task"`
			TaskStatus  string   `json:"taskStatus"`
			TaskOutput  string   `json:"taskOutput"`
			CmdList     []string `json:"cmdList"`
			RetryCnt    int      `json:"retryCnt"`
		}

		// unmarshalling json data
		var t Task
		err := json.Unmarshal([]byte(msg.response),&t)
		if err != nil{
			return m ,nil
		}



		// iterating the cmdList
		var Output string
		for i,cmd := range t.CmdList{
			Output += fmt.Sprintf("Cmd %v --▶ %v\n",i,cmd)

			// logging the cmds in cmdList.txt
			g_logger.Printf("Cmd %v --▶ %v\n" , i,cmd)
		}




		if t.Task == 2{

			if t.TaskStatus == "Completed" {

				m.retryCnt = 0
				m.workingMsg = "Analyzing...⏳"
				m.workingView = false
				m.messages = append(m.messages, m.senderStyle.Render("AI🚀: ")+  yellowColor(t.TaskOutput))
				m.messages = append(m.messages, m.senderStyle.Render("✅  : ")+  greenColor("Task Completed"))
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				return m , nil

			}else{// if task is running or failed


				if m.retryCnt <= config.MaxRetry {
					if m.retryCnt != 0 {
						m.messages = append(m.messages,yellowColor(fmt.Sprintf("Retry %v",m.retryCnt)))
					}
					m.messages = append(m.messages, m.senderStyle.Render("AI🚀: ")+  pinkColor("Command List by Minus1 🤖"))
					m.messages = append(m.messages,greenColor(Output))
					m.viewport.SetContent(strings.Join(m.messages, "\n"))
					m.viewport.GotoBottom()
	
					// pass the tasklist and model to cmdExecutor
					m.workingMsg = "Processing Commands..."
					m.retryCnt += 1 
					return m , cmdExecutor(t.CmdList)
				}else{
					m.messages = append(m.messages,yellowColor("MAX Retry exhausted😫😢😢"))
					m.viewport.SetContent(strings.Join(m.messages, "\n"))
					m.viewport.GotoBottom()
	
					m.workingView = false
					// pass the tasklist and model to cmdExecutor
					// m.workingMsg = "Processing Commands..."
					m.retryCnt = 0 
					return m , nil
				}
			
			}


		}else{

			m.retryCnt = 0
			m.messages = append(m.messages, m.senderStyle.Render("AI🚀: ")+  greenColor(t.TaskOutput))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			m.workingMsg = "Analyzing...⏳"
			m.workingView = false

		}

		return m , nil

	case cmdExecutorMsg:

		if msg.response == "ok"{

			// again ask ai for validity
			m.workingMsg = "Commands Execution finished"
			m.messages = append(m.messages,cyanColor("✅ Execution complete"))
			m.messages = append(m.messages,cyanColor("👀 Awaiting verification"))
			m.messages = append(m.messages,cyanColor("⏳ Please wait\n"))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()


			content, _ := os.ReadFile("cmdLog.txt")
			g_file_2.Truncate(0)
			return m , askAI(fmt.Sprintf("Though everything ran,check the log below for confirmation if the taskStatus completed or still needs update.\n%v",string(content)))




		}else{// if not ok
			m.workingMsg = "🛠️🔍 Entering Error Correction Mode"

			m.messages = append(m.messages,redColor(msg.response))
			m.messages = append(m.messages,yellowColor("🛠️ Entering Error Correction Mode"))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()


			// again ask ai for error correction
			content, _ := os.ReadFile("cmdLog.txt")
			g_file_2.Truncate(0)
			return m , askAI(fmt.Sprintf("%v. Below is the log . Determine whats the problem and generate an updated cmdList \n%v",msg.response,string(content)))

		}


		// return m , nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m model) View() string {

	if m.workingView{
		return fmt.Sprintf(
			"%s%s\n%s --▶ %s \n\n%s\n",
			m.viewport.View(),
			m.helpView(),
			m.spinner.View(),
			m.workingMsg,
			m.textarea.View(),
		) + "\n\n"
	}else{
		return fmt.Sprintf(
			"%s%s\n\n%s\n",
			m.viewport.View(),
			m.helpView(),
			m.textarea.View(),
		) + "\n\n"
	}
}


// 1st session
func askAI(input string) tea.Cmd {
	return func() tea.Msg {

		out := script.AskGemini(&g_client, g_ctx, &g_cs, input)
		// converting from genai.content to string
		data := fmt.Sprintf("%v", out.Parts[0])

		// data := fmt.Sprintf("[\"%s\"]", strings.Join(out.Parts[0], "\", \""))
		// fmt.Println(data)

		return aiMsg{response: data}
	}
}


func cmdExecutor(cmdList []string) tea.Cmd{

	return func() tea.Msg {



		// script.RunCommand(cmdList[0],g_logger_2)

		g_logger_2.Println("--------LOG file START---------")


		for i,cmd := range cmdList{


			g_logger_2.Printf("CMD %v: %s\n", i , cmd)
			s , err := script.RunCommand(cmd , g_logger_2)
			if err != nil {
				g_logger_2.Printf("LOG : %v",err.Error())
				g_logger_2.Printf("ERR : error running cmd %v\n",i)
				g_logger_2.Println("--------LOG file END---------")

				return cmdExecutorMsg{response: "Failed on CMD "+ fmt.Sprint(i) + "\n" }
			}


			// s maybe "ok" but still we need to pass the log to llm for varification

			g_logger_2.Printf("LOG : %v",s)

		}
		// time.Sleep(8 * time.Second) 

		g_logger_2.Println("--------LOG file END---------")

		return cmdExecutorMsg{response: "ok"}
	}
}



