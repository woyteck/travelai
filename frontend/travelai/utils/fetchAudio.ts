export const fetchAudio = async (text: string) => {
    const url = "http://192.168.20.55:8080/talk";
    const response = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ text: text }),
    });

    return await response.blob();
};
