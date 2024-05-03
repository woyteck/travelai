package prompts

import "woyteck.pl/travelai/openai"

func ClassifyQuestion(text string) string {
	context := "Klasyfikuję tekst na kategorie. Zawsze zwracam tylko nazwę kategorii małymi literami bez żadnego formatowania."
	context += "Dostępne kategorie:"
	context += "info - jeśli użytkownik podaje jakieś informacje o sobie, swoich zainteresowaniach, ulubionych rzeczach, marzeniach itp."
	context += "question - jeśli użytkownik zadaje pytanie"

	messages := []openai.Message{
		{Role: "system", Content: context},
		{Role: "user", Content: text},
	}

	completions := openai.GetCompletionShort(messages, "gpt-3.5-turbo")
	if len(completions.Choices) == 0 {
		return ""
	}

	return completions.Choices[0].Message.Content
}
