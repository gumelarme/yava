class Person{
    public String name = "Mark"; 
    public Person(){
    }

    public Person(String name){
        this.name = name;
    }


	public static void main(String[] args){
        Person mark = new Person();
        Person bob = new Person("Bob");

        System.out.println(mark.name);
        System.out.println(bob.name);
	}
}
