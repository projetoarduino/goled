package main


import (
	"fmt"
	"github.com/projetoarduino/goled"
)

var oled = goled.NewDisplayBuffer(128,64)

func main(){
	fmt.Println("Oled Driver")
	
	oled.Init()

	oled.WriteString("projetoarduino.com.br", 0 ,2, 1) // texto , eixo x, eixo y, tamanho da fonte


	//oled.Draw_line(0, 0, 120, 0) 		//Barra horizontal superior
	//oled.Draw_line(120, 0, 120, 60) 	//Barra Direita vertical
	//oled.Draw_line(120, 59, 0, 59) 	//Barra inferior horizontal
	//oled.Draw_line(0, 59, 0, 0); 	 	// Barra esquerda vertical

	// oled.GenIcon(goled.Icon.Cloud, 60, 0) //Constroi um icone


	oled.Display() //Chame essa função toda vez que quiser chamar algo novo
}
