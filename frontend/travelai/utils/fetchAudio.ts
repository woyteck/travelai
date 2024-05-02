export const fetchAudio = async (text: string, conversationId: string) => {
    const url = "http://192.168.1.23:8080/talk";
    let payload: any = { text: text }
    if (conversationId !== "") {
        payload.conversationId = conversationId
    }
    const response = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(payload),
    });

    return response.blob();
};
