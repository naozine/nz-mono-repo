package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create an Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>Simple Top Page</title>
    <meta charset="UTF-8">
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        button { padding: 10px 20px; margin: 10px 0; background-color: #3490dc; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background-color: #2779bd; }
    </style>
</head>
<body>
    <h1>シンプルなトップページ</h1>
    <p>Echo フレームワークを使用したシンプルなトップページです。</p>
    
    <div x-data="{ 
        message: '', 
        clickedGadd9: false,
        clickedA: false,
        clickedFSharpM: false,
        clickedBm: false,
        audioContext: null,
        
        init() {
            // Initialize audio context on first user interaction
        },
        
        async playChord(frequencies, chordName) {
            try {
                // Create audio context if not exists
                if (!this.audioContext) {
                    this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
                }
                
                // Resume audio context if suspended (required by browser policies)
                if (this.audioContext.state === 'suspended') {
                    await this.audioContext.resume();
                }
                
                const duration = 2; // 2 seconds
                
                // Create oscillators for each note
                frequencies.forEach((freq, index) => {
                    const oscillator = this.audioContext.createOscillator();
                    const gainNode = this.audioContext.createGain();
                    
                    oscillator.connect(gainNode);
                    gainNode.connect(this.audioContext.destination);
                    
                    oscillator.frequency.setValueAtTime(freq, this.audioContext.currentTime);
                    oscillator.type = 'sine';
                    
                    // Set volume envelope
                    gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
                    gainNode.gain.linearRampToValueAtTime(0.1, this.audioContext.currentTime + 0.1);
                    gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioContext.currentTime + duration);
                    
                    // Start and stop oscillator
                    oscillator.start(this.audioContext.currentTime);
                    oscillator.stop(this.audioContext.currentTime + duration);
                });
                
                this.message = chordName + 'コードを再生しました！';
                
                // Update clicked state based on chord
                if (chordName === 'Gadd9') this.clickedGadd9 = true;
                else if (chordName === 'A') this.clickedA = true;
                else if (chordName === 'F#m') this.clickedFSharpM = true;
                else if (chordName === 'Bm') this.clickedBm = true;
                
            } catch (error) {
                console.error('Web Audio API error:', error);
                alert('音声の再生に失敗しました。ブラウザがWeb Audio APIに対応していない可能性があります。');
            }
        },
        
        async playGadd9() {
            // Gadd9 chord frequencies (G-B-D-A)
            // G4 = 392 Hz, B4 = 494 Hz, D5 = 587 Hz, A4 = 440 Hz
            const frequencies = [392, 494, 587, 440];
            await this.playChord(frequencies, 'Gadd9');
        },
        
        async playA() {
            // A chord frequencies (A-C#-E)
            // A4 = 440 Hz, C#5 = 554 Hz, E5 = 659 Hz
            const frequencies = [440, 554, 659];
            await this.playChord(frequencies, 'A');
        },
        
        async playFSharpMinor() {
            // F#m chord frequencies (F#-A-C#)
            // F#4 = 370 Hz, A4 = 440 Hz, C#5 = 554 Hz
            const frequencies = [370, 440, 554];
            await this.playChord(frequencies, 'F#m');
        },
        
        async playBMinor() {
            // Bm chord frequencies (B-D-F#)
            // B4 = 494 Hz, D5 = 587 Hz, F#5 = 740 Hz
            const frequencies = [494, 587, 740];
            await this.playChord(frequencies, 'Bm');
        }
    }">
        <button @click="playGadd9()" x-text="clickedGadd9 ? 'Gadd9再生済み' : 'Gadd9コードを再生'"></button>
        <button @click="playA()" x-text="clickedA ? 'A再生済み' : 'Aコード（ラ・ド#・ミ）を再生'"></button>
        <button @click="playFSharpMinor()" x-text="clickedFSharpM ? 'F#m再生済み' : 'F#mコード（ファ#・ラ・ド#）を再生'"></button>
        <button @click="playBMinor()" x-text="clickedBm ? 'Bm再生済み' : 'Bmコード（シ・レ・ファ#）を再生'"></button>
        <p x-show="message" x-text="message"></p>
    </div>
</body>
</html>`)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
