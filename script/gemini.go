package script

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
)

func AskGemini(client *genai.Client , ctx context.Context ,cs *genai.ChatSession, text string) genai.Content {

	send := func(msg string) *genai.GenerateContentResponse {
		// fmt.Printf("== Me: %s\n== Model:\n", msg)
		res, err := cs.SendMessage(ctx, genai.Text(msg))
		if err != nil {
			log.Fatal(err)
		}
		return res
	}
	res := send(text)
	
	// Handle the response of generated text.
    for _, c := range res.Candidates {
        if c.Content != nil {
            return *c.Content
        }
    }
	return genai.Content{}
	// for i, c := range cs.History {
	// 	log.Printf("    %d: %+v", i, c)
	// }
	
}

func CloseGemini(client *genai.Client){
	defer client.Close()
}

