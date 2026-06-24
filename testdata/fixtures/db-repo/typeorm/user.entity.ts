import { Column, Entity, ManyToOne, OneToMany, PrimaryGeneratedColumn } from "typeorm";

@Entity("users")
export class UserEntity {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ unique: true })
  email: string;

  @OneToMany(() => OrderEntity, order => order.user)
  orders: OrderEntity[];
}

@Entity("orders")
export class OrderEntity {
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  total: number;

  @ManyToOne(() => UserEntity, user => user.orders)
  user: UserEntity;
}
