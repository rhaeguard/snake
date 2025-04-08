# retro-style snake game 

written in [raylib](https://github.com/raysan5/raylib) (using [raylib-go](https://github.com/gen2brain/raylib-go))

the gameplay

https://github.com/user-attachments/assets/e437530c-a5b4-4204-b25a-6a67be830c45

- for each game, a new map is procedurally generated using [Wave Function Collapse](https://robertheaton.com/2018/12/17/wavefunction-collapse-algorithm/) algorithm.

## build & run

```sh
# in the project root
# to build:
go build -o bin\ -ldflags "-H=windowsgui -s -w" .
# to run
go run .
```

There's a Windows executable file already in the [`bin`](./bin/) folder. 

## todo

- [x] basic movement mechanics
- [x] eating food grows the snake
- [x] hitting border kills
- [x] eating self kills
- [x] ephemeral food
- [x] display:
  - [x] difficulty
  - [x] score
  - [x] max score
  - [x] game title
- [x] intro menu
  - [x] current high score
  - [x] difficulty (slug, worm, python)
  - [x] game logo?
- [x] procedurally generated map
- [ ] different types of items
  - [ ] poison
- [ ] music?

Reference image used so far (credit: https://metro.co.uk):

![](https://metro.co.uk/wp-content/uploads/2015/05/snake_mobile.gif)

## Font

I am using the Minecraft font which is 100% free, you can find the Minecraft font here: https://www.dafont.com/minecraft.font