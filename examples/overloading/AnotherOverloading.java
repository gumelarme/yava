class AnotherOverload{
    public String greet(){
        return "Good morning!";
    }

    public void greet (String name){
        System.out.println("Hello, ");
        System.out.println(name);
    }

	public static void main(String[] args){
		AnotherOverload ov =  new AnotherOverload();
		System.out.println(ov.greet());
        ov.greet("Mark");
	}
}
