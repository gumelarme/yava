class Fib{
	public int fibonacci(int n)  {
		if (n < 2) return n;
		return this.fibonacci(n - 1) + this.fibonacci(n - 2);
	}

	public static void main(String[] args){
		Fib fib = new Fib();
		int result = fib.fibonacci(9);

		// 1, 1, 2, 3, 5, 8, 13, 21, 34
		System.out.println(result);
	}
}
