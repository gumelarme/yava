interface ISpeak{
    public void speak();
}

class Human implements ISpeak {
    public String name;
    public Human(String name){
        this.name = name; 
    }
    public void speak(){
        System.out.println(this.name);
    }
}

class Cat implements ISpeak {
    public void speak(){
        System.out.println("Meoow!");
    }
}

class Main{
    public void callSpeaker(ISpeak object){
        object.speak();
    }

    public static void main(String[] args){
        Human bob = new Human("Bob");
        Cat cat = new Cat();

        Main main = new Main();
        main.callSpeaker(bob);
        main.callSpeaker(cat);
    }
}
