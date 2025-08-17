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
        message: 'Gadd9コードを再生しました！', 
        clicked: false,
        audioContext: null,
        
        init() {
            // Initialize audio context on first user interaction
        },
        
        async playGadd9() {
            try {
                // Create audio context if not exists
                if (!this.audioContext) {
                    this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
                }
                
                // Resume audio context if suspended (required by browser policies)
                if (this.audioContext.state === 'suspended') {
                    await this.audioContext.resume();
                }
                
                // Gadd9 chord frequencies (G-B-D-A)
                // G4 = 392 Hz, B4 = 494 Hz, D5 = 587 Hz, A4 = 440 Hz
                const frequencies = [392, 494, 587, 440];
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
                
                this.clicked = true;
            } catch (error) {
                console.error('Web Audio API error:', error);
                alert('音声の再生に失敗しました。ブラウザがWeb Audio APIに対応していない可能性があります。');
            }
        }
    }">
        <button @click="playGadd9()" x-text="clicked ? 'Gadd9再生済み' : 'Gadd9コードを再生'"></button>
        <p x-show="clicked" x-text="message"></p>
    </div>
</body>
</html>`)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
