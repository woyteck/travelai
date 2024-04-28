const fetchAudio = async (text: string) => {
    const response = await fetch(process.env.BACKEND_URL!, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ text: text }),
    });

    return await response.blob();
};
