class FizzBuzz{
	public static void main(String[] args){
		for(int a = 1; a < 20; a = a + 1){
			if(a % 15 == 0){
				System.out.println("FizzBuzz");
			}if(a % 3 == 0){
				System.out.println("Fizz");
			} else if(a % 5 ==0){
				System.out.println("Buzz");
			}else {
				System.out.println(a);
			}
		}
	}
}
