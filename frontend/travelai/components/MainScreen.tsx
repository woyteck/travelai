import { useState } from "react";
import { Button, Pressable, StyleSheet, Text, View } from "react-native";
import { useVoiceRecognition } from "../hooks/useVoiceRecognition";
import * as FileSystem from "expo-file-system"
import {Audio} from "expo-av";
import { writeAudioToFile } from "../utils/writeAudioToFile";
import { playFromPath } from "../utils/playFromPath";
import { fetchAudio } from "../utils/fetchAudio";

Audio.setAudioModeAsync({
  allowsRecordingIOS: false,
  staysActiveInBackground: false,
  playsInSilentModeIOS: true,
  shouldDuckAndroid: true,
  playThroughEarpieceAndroid: false,
});

export default function MainScreen() {
  const [borderColor, setBorderColor] = useState<"lightgray" | "lightgreen">("lightgray");
  const {state, startRecognizing, stopRecognizing, destroyRecognizer} = useVoiceRecognition();
  const [urlPath, setUrlPath] = useState("");

  const listFiles = async () => {
    try {
      const result = await FileSystem.readDirectoryAsync(FileSystem.documentDirectory!);
      if (result.length > 0) {
        const fileName = result[0];
        const path = FileSystem.documentDirectory + fileName;
        setUrlPath(path);
      }
    } catch(e) {
      console.log(e);
    }
  }

  const handleSubmit = async () => {
    if (!state.results[0]) {
      return;
    }

    try {
      const audioBlob = await fetchAudio(state.results[0]);
      const reader = new FileReader();
      reader.onload = async (e) => {
        if (e.target && typeof e.target.result === "string") {
          const audioData = e.target.result.split(",")[1];
          const path = await writeAudioToFile(audioData);
          setUrlPath(path)
          await playFromPath(path);
          destroyRecognizer();
        }
      };
      reader.readAsDataURL(audioBlob);
    } catch(err) {
      console.log(err);
    }
  }

  return (
    <View style={styles.container}>
      <Text style={{ fontSize: 32, fontWeight: "bold", marginBottom: 30 }}>Talk to me</Text>
      <Text>Press and hold this button to speak</Text>
      
      <Text style={styles.message}>Your message:</Text>
      <Pressable 
        style={[styles.pressable, {borderColor: borderColor}]}
        onPressIn={() => {
            setBorderColor("lightgreen");
            startRecognizing();
        }}
        onPressOut={() => {
            setBorderColor("lightgray");
            stopRecognizing();
            handleSubmit();
        }}
      >
        <Text>Hold to speak</Text>
      </Pressable>
      <Button title="Replay last message" onPress={async () => {
        await playFromPath(urlPath);
      }} />
      <Text>{JSON.stringify(state, null, 2)}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  message: {
    textAlign: 'center',
    color: "#0000ff",
    marginVertical: 10,
    fontSize: 17,
    fontWeight: "bold",
  },
  pressable: {
    width: "90%",
    padding: 30,
    gap: 10,
    borderWidth: 3,
    alignItems: "center",
    borderRadius: 10,
  },
});

  