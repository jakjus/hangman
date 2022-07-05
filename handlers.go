package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"unicode"
)

type game struct {
	inProgress bool
	failed     int
	answer     []rune
	guessed    []rune
	channelId  string
	msgId      string
	session    *discordgo.Session
}

func containsChar(arr []rune, char rune) bool {
	for _, el := range arr {
		if el == char {
			return true
		}
	}
	return false
}

func (g *game) gameState() string {
	endInfo := ""
	if !g.inProgress {
		endInfo = "Game Finished!"
	}
	return fmt.Sprintf("guessed: %s, answer: %s, failed: %v\n%v", string(g.guessed), string(g.answer), g.failed, endInfo)
}

func (g *game) move(m rune) string {
	if containsChar(g.guessed, m) {
		return "Already guessed."
	}
	if containsChar(g.answer, m) {
		g.guessed = append(g.guessed, m)
		return "Nice!"
	}
	g.failed += 1
	g.guessed = append(g.guessed, m)
	return "Wrong."
}

func shuffleWord() []rune {
	return []rune("phone")
}

func maskWord(word []rune, guessed []rune) []rune {
	for i := 0; i < len(word); i++ {
    shouldReveal := false
    for _, guess := range guessed {
      if word[i] == guess {
        shouldReveal = true
      }
    }
		if !shouldReveal && (word[i] != ' ') {
			word[i] = '_'
		}
	}
	return word
}

func (g *game) start(s *discordgo.Session, i *discordgo.InteractionCreate) {
	word := shuffleWord()
	masked := maskWord(word, g.guessed)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Game Started. `Word: %v`", masked),
		},
	})

	newGame := &game{
		inProgress: true,
		channelId:  i.Interaction.ChannelID,
		session:    s,
	}

	games[i.Interaction.Member.User.ID] = newGame
}

func (g *game) checkIfEnd(i *discordgo.MessageCreate) {
	if g.failed == 7 {
		g.inProgress = false
		games[i.Interaction.Member.User.ID] = nil
	}
}

func (g *game) send(msg string) (*discordgo.Message, error) {
	var ackMsg *discordgo.Message
	var err error
	if g.msgId == "" {
		ackMsg, err = g.session.ChannelMessageSend(g.channelId, msg)
	} else {
		ackMsg, err = g.session.ChannelMessageEdit(g.channelId, g.msgId, msg)
	}
	return ackMsg, err
}

var (
	games           map[string]*game = make(map[string]*game)
	commandHandlers                  = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if g, ok := games[i.Interaction.Member.User.ID]; ok {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You have already game in progress! " + g.gameState(),
					},
				})
			} else {
				g.start(s, i)
			}
		},
	}

	msgHandler = func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if (len(i.Message.Content) != 1) || !unicode.IsLetter(rune(i.Message.Content[0])) {
			return
		}
		if g, ok := games[i.Message.Author.ID]; ok {
			msg := g.move(rune(i.Message.Content[0]))
			_, err := g.send(msg)
			if err != nil {
				return
			}
			err = s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
			if err != nil {
				g.send("Error: I do not have privileges to delete user's message (required).")
				return
			}
			g.send(g.gameState())
			g.checkIfEnd(i)
		}
	}
)
