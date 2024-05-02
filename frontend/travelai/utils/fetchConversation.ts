export const fetchConversation = async () => {
    let url = "http://192.168.1.23:8080/conversation";
    // if (conversationId) {
    //     url += `/${conversationId}`;
    // }
    
    const response = await fetch(url, {
        method: "GET",
    });

    return response.json();
};
